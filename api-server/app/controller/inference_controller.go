package controller

import (
	"aitrics-vital-signs/api-server/domain/inference"
	"aitrics-vital-signs/api-server/internal/output"
	pkgError "aitrics-vital-signs/library/error"

	"github.com/gin-gonic/gin"
)

type inferenceController struct {
	service inference.InferenceService
}

// CalculateVitalRisk
// @Security Bearer
// @Title CalculateVitalRisk
// @Description Vital 데이터 기반 위험 스코어 계산
// @Tags V1 - Inference
// @Accept json
// @Produce json
// @Param reqBody body inference.VitalRiskRequest true "위험 스코어 계산 요청"
// @Success 200 {object} output.Output{data=inference.VitalRiskResponse}
// @Failure 400 {object} output.Output "code: 400001 - Wrong parameter"
// @Failure 500 {object} output.Output "code: 100003 - Fail to get data from db"
// @Router /v1/inference/vital-risk [Post]
func (i *inferenceController) CalculateVitalRisk(ctx *gin.Context) {
	var reqBody inference.VitalRiskRequest
	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		output.AppendErrorContext(ctx, pkgError.WrapWithCode(err, pkgError.WrongParam, err.Error(), "fail to parse request parameter"), nil)
		return
	}

	result, err := i.service.CalculateVitalRisk(ctx, reqBody)
	if err != nil {
		output.AppendErrorContext(ctx, pkgError.Wrap(err), nil)
		return
	}

	output.Send(ctx, result)
}

func NewInferenceController(service inference.InferenceService) inference.InferenceController {
	return &inferenceController{
		service: service,
	}
}
