package main

import (
	"aitrics-vital-signs/api-server/app/controller"
	"aitrics-vital-signs/api-server/app/external"
	"aitrics-vital-signs/api-server/app/repository"
	"aitrics-vital-signs/api-server/app/router"
	"aitrics-vital-signs/api-server/app/service"
	"aitrics-vital-signs/api-server/internal/middleware"
	"aitrics-vital-signs/library/envs"
	pkgLogger "aitrics-vital-signs/library/logger"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "aitrics-vital-signs/api-server/docs"
)

// @title Vital Signs API
// @version 1.0
// @description Vital Signs Backend API
// @host localhost:8080
// @BasePath /api

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
func main() {
	pkgLogger.MustInitZapLogger()
	if pkgLogger.ZapLogger == nil {
		log.Fatal("logger is nil")
	}

	bCtx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	group, _ := errgroup.WithContext(bCtx)

	engine := gin.New()
	engine.Use(gin.Recovery(), middleware.GinBusinessErrLogger())

	conf := &cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS", "UPDATE"},
		AllowHeaders:     []string{"X-Request-Id", "X-Forwarded-Proto", "X-Forwarded-Host", "Origin", "Content-Length", "Access-Control-Allow-Origin", "Content-Type", "Accept-Encoding", "origin", "accept", "X-Requested-With", " X-CSRF-Token", "Cache-Control", "Baggage"},
		AllowCredentials: false,
		ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Headers", "Cache-Control", "Content-Language", "Content-Type"},
		MaxAge:           12 * time.Hour,
		AllowOrigins:     []string{"*"},
	}
	engine.Use(cors.New(*conf))
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	dbClient := external.MustExternalDB()

	patientRepository := repository.NewPatientRepository(dbClient)
	vitalRepository := repository.NewVitalRepository(dbClient)

	patientService := service.NewPatientService(patientRepository, vitalRepository)
	vitalService := service.NewVitalService(vitalRepository, patientRepository)
	inferenceService := service.NewInferenceService(vitalRepository, patientRepository)

	patientController := controller.NewPatientController(patientService)
	vitalController := controller.NewVitalController(vitalService)
	inferenceController := controller.NewInferenceController(inferenceService)

	router.NewPatientRouter(engine, patientController)
	router.NewVitalRouter(engine, vitalController)
	router.NewInferenceRouter(engine, inferenceController)

	s := &http.Server{
		Addr:    fmt.Sprintf(":%s", envs.ServerPort),
		Handler: engine,
	}

	group.Go(func() error {
		err := s.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			pkgLogger.ZapLogger.Logger.Info("server closed gracefully")
			return nil
		}
		return err
	})

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	defer close(interrupt)

	select {
	case <-interrupt:
		pkgLogger.ZapLogger.Logger.Info("received shutdown signal")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.Shutdown(ctx); err != nil {
			pkgLogger.ZapLogger.Logger.Error("server shutdown failed: " + err.Error())
		} else {
			pkgLogger.ZapLogger.Logger.Info("server gracefully stopped")
		}
	}

	if err := group.Wait(); err != nil {
		pkgLogger.ZapLogger.Logger.Fatal(err.Error())
	}

	pkgLogger.ZapLogger.Logger.Info("API Server End")
}
