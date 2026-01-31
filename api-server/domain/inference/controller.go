//go:generate mockgen -source=controller.go -destination=../mock/mock_inference_controller.go -package=mock
package inference

import "github.com/gin-gonic/gin"

type InferenceController interface {
	CalculateVitalRisk(ctx *gin.Context)
}