package controller

import (
	"aitrics-vital-signs/api-server/domain/patient"
	"aitrics-vital-signs/api-server/internal/output"
	pkgError "aitrics-vital-signs/library/error"

	"github.com/gin-gonic/gin"
)

type patientController struct {
	service patient.PatientService
}

// CreatePatient
// @Title CreatePatient
// @Description 환자 등록
// @Tags V1 - Patient
// @Accept json
// @Produce json
// @Param reqBody body patient.CreatePatientRequest true "execute hook request"
// @Success 200 {object} output.Output
// @Failure 400 {object} output.Output "code: 400001 - Wrong parameter"
// @Failure 500 {object} output.Output "code: 100001 - Fail to create data from db"
// @Router /v1/patients [Post]
func (p *patientController) CreatePatient(ctx *gin.Context) {
	var reqBody patient.CreatePatientRequest
	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		output.AppendErrorContext(ctx, pkgError.WrapWithCode(err, pkgError.WrongParam, err.Error(), "fail to parse request parameter"), nil)
		return
	}

	if err := p.service.CreatePatient(ctx, reqBody); err != nil {
		output.AppendErrorContext(ctx, pkgError.Wrap(err), nil)
		return
	}

	output.Send(ctx, nil)
}

// UpdatePatient
// @Title UpdatePatient
// @Description 환자 정보 수정
// @Tags V1 - Patient
// @Accept json
// @Produce json
// @Param patient_id path string true "환자 ID"
// @Param reqBody body patient.UpdatePatientRequest true "환자 정보 수정 요청"
// @Success 200 {object} output.Output
// @Failure 400 {object} output.Output "code: 400001 - Wrong parameter"
// @Failure 409 {object} output.Output "code: 400002 - Version conflict"
// @Failure 500 {object} output.Output "code: 100002 - Fail to update data from db"
// @Router /v1/patients/{patient_id} [Put]
func (p *patientController) UpdatePatient(ctx *gin.Context) {
	patientID := ctx.Param("patient_id")
	if patientID == "" {
		output.AppendErrorContext(ctx, pkgError.WrapWithCode(pkgError.EmptyBusinessError(), pkgError.WrongParam, "patient_id is required"), nil)
		return
	}

	var reqBody patient.UpdatePatientRequest
	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		output.AppendErrorContext(ctx, pkgError.WrapWithCode(err, pkgError.WrongParam, err.Error(), "fail to parse request parameter"), nil)
		return
	}

	if err := p.service.UpdatePatient(ctx, patientID, reqBody); err != nil {
		output.AppendErrorContext(ctx, pkgError.Wrap(err), nil)
		return
	}

	output.Send(ctx, nil)
}

func NewPatientController(service patient.PatientService) patient.PatientController {
	p := &patientController{
		service: service,
	}

	return p
}
