package router

import (
	"aitrics-vital-signs/api-server/domain/vital"
	"aitrics-vital-signs/api-server/internal/middleware"

	"github.com/gin-gonic/gin"
)

func NewVitalRouter(engine *gin.Engine, controller vital.VitalController) {
	v1Group := engine.Group("/api/v1")
	v1Group.Use(middleware.ValidTokenMiddleware())

	vitalGroup := v1Group.Group("/vitals")
	{
		vitalGroup.POST("", controller.UpsertVital)
	}
}
