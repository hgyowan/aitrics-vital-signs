package repository

import (
	"aitrics-vital-signs/api-server/domain"
	"aitrics-vital-signs/api-server/domain/vital"
	pkgError "aitrics-vital-signs/library/error"
	"context"
	"time"
)

type vitalRepository struct {
	externalGormClient domain.ExternalDBClient
}

func (v *vitalRepository) FindVitalByPatientIDAndRecordedAtAndVitalType(ctx context.Context, patientID string, recordedAt time.Time, vitalType string) (*vital.Vital, error) {
	var result vital.Vital
	if err := v.externalGormClient.MySQL().WithContext(ctx).
		Where("patient_id = ? AND recorded_at = ? AND vital_type = ?", patientID, recordedAt, vitalType).
		First(&result).Error; err != nil {
		return nil, pkgError.Wrap(err)
	}
	return &result, nil
}

func (v *vitalRepository) FindVitalsByPatientIDAndDateRange(ctx context.Context, patientID string, from time.Time, to time.Time, vitalType string) ([]vital.Vital, error) {
	var results []vital.Vital
	query := v.externalGormClient.MySQL().WithContext(ctx).
		Where("patient_id = ? AND recorded_at >= ? AND recorded_at <= ?", patientID, from, to)

	// vitalType이 있으면 해당 타입만 필터링
	if vitalType != "" {
		query = query.Where("vital_type = ?", vitalType)
	}

	if err := query.Order("recorded_at ASC").Find(&results).Error; err != nil {
		return nil, pkgError.Wrap(err)
	}

	return results, nil
}

func (v *vitalRepository) CreateVital(ctx context.Context, model *vital.Vital) error {
	return pkgError.Wrap(v.externalGormClient.MySQL().WithContext(ctx).Create(model).Error)
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
		return pkgError.Wrap(result.Error)
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
