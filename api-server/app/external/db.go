package external

import (
	"aitrics-vital-signs/api-server/domain"
	"aitrics-vital-signs/library/envs"
	pkgLogger "aitrics-vital-signs/library/logger"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	maxOpenConnNum  = 25
	maxIdleConnNum  = 20
	connMaxLifetime = 5 * time.Minute
)

type externalDB struct {
	mysql *gorm.DB
}

func (e *externalDB) MySQL() *gorm.DB {
	return e.mysql
}

func MustExternalDB() domain.ExternalDBClient {
	// Create the DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=UTC",
		envs.DBUser, envs.DBPassword, envs.DBHost, envs.DBPort, envs.DBName, // Database Info
	)
	// Initialize GORM DB connection
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
		Logger:                 pkgLogger.ZapLogger.GormLogger,
	})
	if err != nil {
		pkgLogger.ZapLogger.Logger.Sugar().Fatalf("failed to connect to database: %v", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		pkgLogger.ZapLogger.Logger.Sugar().Fatalf("failed to get database object: %v", err)
	}
	sqlDB.SetMaxOpenConns(maxOpenConnNum)
	sqlDB.SetMaxIdleConns(maxIdleConnNum)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)

	pkgLogger.ZapLogger.Logger.Info("MySQL connection established successfully!")

	return &externalDB{mysql: db}
}
