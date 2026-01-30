//go:generate mockgen -source=controller.go -destination=../mock/mock_controller.go -package=mock
package patient

import "github.com/gin-gonic/gin"

type PatientController interface {
	CreatePatient(ctx *gin.Context)
	UpdatePatient(ctx *gin.Context)
}
