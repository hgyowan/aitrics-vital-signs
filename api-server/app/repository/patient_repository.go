package repository

import (
	"aitrics-vital-signs/api-server/domain"
	"aitrics-vital-signs/api-server/domain/patient"
	pkgError "aitrics-vital-signs/library/error"
	"context"
)

type patientRepository struct {
	externalGormClient domain.ExternalDBClient
}

func (p *patientRepository) CreatePatient(ctx context.Context, model *patient.Patient) error {
	return pkgError.WrapWithCode(p.externalGormClient.MySQL().WithContext(ctx).Create(model).Error, pkgError.Create)
}

func (p *patientRepository) FindPatientByID(ctx context.Context, patientID string) (*patient.Patient, error) {
	var result patient.Patient
	if err := p.externalGormClient.MySQL().WithContext(ctx).
		Where("patient_id = ?", patientID).
		First(&result).Error; err != nil {
		return nil, pkgError.WrapWithCode(err, pkgError.Get)
	}
	return &result, nil
}

func (p *patientRepository) UpdatePatient(ctx context.Context, model *patient.Patient) error {
	// Optimistic Lock: WHERE version = (oldVersion) 조건으로 업데이트
	// version은 이미 Service layer에서 +1 증가된 상태
	oldVersion := model.Version - 1

	result := p.externalGormClient.MySQL().WithContext(ctx).
		Model(&patient.Patient{}).
		Where("id = ? AND version = ?", model.ID, oldVersion).
		Updates(map[string]interface{}{
			"patient_id": model.PatientID,
			"name":       model.Name,
			"gender":     model.Gender,
			"birth_date": model.BirthDate,
			"version":    model.Version,
			"updated_at": model.UpdatedAt,
		})

	if result.Error != nil {
		return pkgError.WrapWithCode(result.Error, pkgError.Update)
	}

	// RowsAffected가 0이면 version conflict
	if result.RowsAffected == 0 {
		return pkgError.WrapWithCode(pkgError.EmptyBusinessError(), pkgError.Conflict, "version conflict in db update")
	}

	return nil
}

func NewPatientRepository(externalGormClient domain.ExternalDBClient) patient.PatientRepository {
	return &patientRepository{externalGormClient: externalGormClient}
}
