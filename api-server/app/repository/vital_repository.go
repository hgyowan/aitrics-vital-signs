package repository

import (
	"aitrics-vital-signs/api-server/domain"
	"aitrics-vital-signs/api-server/domain/vital"
	pkgError "aitrics-vital-signs/library/error"
	"context"
	"errors"

	"gorm.io/gorm"
)

type vitalRepository struct {
	externalGormClient domain.ExternalDBClient
}

func (v *vitalRepository) FindVitalByPatientIDAndRecordedAtAndVitalType(ctx context.Context, param vital.FindVitalByPatientIDAndRecordedAtAndVitalTypeParam) (*vital.Vital, error) {
	var result vital.Vital
	if err := v.externalGormClient.MySQL().WithContext(ctx).
		Where("patient_id = ? AND recorded_at = ? AND vital_type = ?", param.PatientID, param.RecordedAt, param.VitalType).
		First(&result).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, pkgError.WrapWithCode(err, pkgError.NotFound)
		}
		return nil, pkgError.WrapWithCode(err, pkgError.Get)
	}
	return &result, nil
}

func (v *vitalRepository) FindVitalsByPatientIDAndDateRange(ctx context.Context, param vital.FindVitalsByPatientIDAndDateRangeParam) ([]vital.Vital, error) {
	var results []vital.Vital
	query := v.externalGormClient.MySQL().WithContext(ctx).
		Where("patient_id = ? AND recorded_at >= ? AND recorded_at <= ?", param.PatientID, param.From, param.To)

	// vitalType이 있으면 해당 타입만 필터링
	if param.VitalType != "" {
		query = query.Where("vital_type = ?", param.VitalType)
	}

	if err := query.Order("recorded_at DESC").Find(&results).Error; err != nil {
		return nil, pkgError.WrapWithCode(err, pkgError.Get)
	}

	return results, nil
}

func (v *vitalRepository) CreateVital(ctx context.Context, model *vital.Vital) error {
	return pkgError.WrapWithCode(v.externalGormClient.MySQL().WithContext(ctx).Create(model).Error, pkgError.Create)
}

func (v *vitalRepository) UpdateVital(ctx context.Context, model *vital.Vital) error {
	// Optimistic Lock: WHERE version = (oldVersion) 조건으로 업데이트
	oldVersion := model.Version - 1

	result := v.externalGormClient.MySQL().WithContext(ctx).
		Model(&vital.Vital{}).
		Where("patient_id = ? AND recorded_at = ? AND vital_type = ? AND version = ?",
			model.PatientID, model.RecordedAt, model.VitalType, oldVersion).
		Updates(map[string]interface{}{
			"value":      model.Value,
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

func NewVitalRepository(externalGormClient domain.ExternalDBClient) vital.VitalRepository {
	return &vitalRepository{externalGormClient: externalGormClient}
}
