package service

import (
	"aitrics-vital-signs/api-server/domain/vital"
	pkgError "aitrics-vital-signs/library/error"
	"context"
	"time"

	"gorm.io/gorm"
)

type vitalService struct {
	repo vital.VitalRepository
}

func (v *vitalService) UpsertVital(ctx context.Context, request vital.UpsertVitalRequest) error {
	// 기존 Vital 데이터 조회
	existingVital, err := v.repo.FindVitalByPatientIDAndRecordedAtAndVitalType(
		ctx,
		request.PatientID,
		request.RecordedAt,
		request.VitalType,
	)

	now := time.Now().UTC()

	// 존재하지 않으면 INSERT
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() || pkgError.CompareBusinessError(err, pkgError.Get) {
			// INSERT: version은 1부터 시작
			if request.Version != 1 {
				return pkgError.WrapWithCode(pkgError.EmptyBusinessError(), pkgError.WrongParam, "version must be 1 for new record")
			}

			if err := v.repo.CreateVital(ctx, &vital.Vital{
				PatientID:  request.PatientID,
				RecordedAt: request.RecordedAt,
				VitalType:  request.VitalType,
				Value:      request.Value,
				Version:    1,
				CreatedAt:  now,
				UpdatedAt:  &now,
			}); err != nil {
				return pkgError.WrapWithCode(err, pkgError.Create)
			}

			return nil
		}

		// 다른 에러는 그대로 반환
		return pkgError.WrapWithCode(err, pkgError.Get)
	}

	// 존재하면 UPDATE (Optimistic Lock 적용)
	if existingVital.Version != request.Version {
		return pkgError.WrapWithCode(pkgError.EmptyBusinessError(), pkgError.Conflict, "version mismatch")
	}

	existingVital.Value = request.Value
	existingVital.Version = request.Version + 1
	existingVital.UpdatedAt = &now

	if err := v.repo.UpdateVital(ctx, existingVital); err != nil {
		// Repository에서 이미 Conflict 에러를 반환하는 경우 그대로 전달
		if pkgError.CompareBusinessError(err, pkgError.Conflict) {
			return err
		}
		return pkgError.WrapWithCode(err, pkgError.Update)
	}

	return nil
}

func (v *vitalService) GetVitalsByPatientIDAndDateRange(ctx context.Context, request vital.GetVitalsRequest) (*vital.GetVitalsResponse, error) {
	// Repository에서 Vital 데이터 조회
	vitals, err := v.repo.FindVitalsByPatientIDAndDateRange(
		ctx,
		request.PatientID,
		request.From,
		request.To,
		request.VitalType,
	)
	if err != nil {
		return nil, pkgError.WrapWithCode(err, pkgError.Get)
	}

	// Response 변환
	items := make([]vital.VitalItemResponse, 0, len(vitals))
	for _, v := range vitals {
		items = append(items, vital.VitalItemResponse{
			VitalType:  v.VitalType,
			RecordedAt: v.RecordedAt,
			Value:      v.Value,
		})
	}

	return &vital.GetVitalsResponse{
		PatientID: request.PatientID,
		Items:     items,
	}, nil
}

func NewVitalService(repo vital.VitalRepository) vital.VitalService {
	return &vitalService{repo}
}
