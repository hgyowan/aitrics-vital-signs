//go:generate mockgen -source=controller.go -destination=../mock/mock_vital_controller.go -package=mock
package vital

import "github.com/gin-gonic/gin"

type VitalController interface {
	UpsertVital(ctx *gin.Context)
}