package controller

import (
	"aitrics-vital-signs/api-server/domain/vital"
	"aitrics-vital-signs/api-server/internal/output"
	pkgError "aitrics-vital-signs/library/error"

	"github.com/gin-gonic/gin"
)

type vitalController struct {
	service vital.VitalService
}

// UpsertVital
// @Security Bearer
// @Title UpsertVital
// @Description Vital 데이터 저장/수정 (UPSERT, Optimistic Lock 적용)
// @Tags V1 - Vital
// @Accept json
// @Produce json
// @Param reqBody body vital.UpsertVitalRequest true "Vital 데이터 저장/수정 요청"
// @Success 200 {object} output.Output
// @Failure 400 {object} output.Output "code: 400001 - Wrong parameter"
// @Failure 409 {object} output.Output "code: 400002 - Version conflict"
// @Failure 500 {object} output.Output "code: 100001 - Fail to create data / code: 100002 - Fail to update data"
// @Router /v1/vitals [Post]
func (v *vitalController) UpsertVital(ctx *gin.Context) {
	var reqBody vital.UpsertVitalRequest
	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		output.AppendErrorContext(ctx, pkgError.WrapWithCode(err, pkgError.WrongParam, err.Error(), "fail to parse request parameter"), nil)
		return
	}

	if err := v.service.UpsertVital(ctx, reqBody); err != nil {
		output.AppendErrorContext(ctx, pkgError.Wrap(err), nil)
		return
	}

	output.Send(ctx, nil)
}

func NewVitalController(service vital.VitalService) vital.VitalController {
	v := &vitalController{
		service: service,
	}

	return v
}
