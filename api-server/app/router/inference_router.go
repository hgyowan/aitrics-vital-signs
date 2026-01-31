package router

import (
	"aitrics-vital-signs/api-server/domain/inference"
	"aitrics-vital-signs/api-server/internal/middleware"

	"github.com/gin-gonic/gin"
)

func NewInferenceRouter(engine *gin.Engine, controller inference.InferenceController) {
	v1Group := engine.Group("/api/v1")
	v1Group.Use(middleware.ValidTokenMiddleware())

	inferenceGroup := v1Group.Group("/inference")
	{
		inferenceGroup.POST("/vital-risk", controller.CalculateVitalRisk)
	}
}
