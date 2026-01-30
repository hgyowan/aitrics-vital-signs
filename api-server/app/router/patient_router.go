package router

import (
	"aitrics-vital-signs/api-server/domain/patient"
	"aitrics-vital-signs/api-server/internal/middleware"

	"github.com/gin-gonic/gin"
)

func NewPatientRouter(engine *gin.Engine, controller patient.PatientController) {
	v1Group := engine.Group("/v1")
	v1Group.Use(middleware.ValidTokenMiddleware())

	patientGroup := v1Group.Group("/patients")
	{
		patientGroup.POST("", controller.CreatePatient)
	}
}
