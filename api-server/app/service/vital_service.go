package service

import (
	"aitrics-vital-signs/api-server/domain/vital"
	pkgError "aitrics-vital-signs/library/error"
	"context"
	"time"
)

type vitalService struct {
	repo vital.VitalRepository
}

func (v *vitalService) UpsertVital(ctx context.Context, request vital.UpsertVitalRequest) error {
	// 기존 Vital 데이터 조회
	existingVital, err := v.repo.FindVitalByPatientIDAndRecordedAtAndVitalType(ctx, vital.FindVitalByPatientIDAndRecordedAtAndVitalTypeParam{
		PatientID:  request.PatientID,
		RecordedAt: request.RecordedAt,
		VitalType:  request.VitalType,
	})

	now := time.Now().UTC()

	// 존재하지 않으면 INSERT
	if err != nil {
		if pkgError.CompareBusinessError(err, pkgError.NotFound) {
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
				return pkgError.Wrap(err)
			}

			return nil
		}

		// 다른 에러는 그대로 반환
		return pkgError.Wrap(err)
	}

	// 존재하면 UPDATE (Optimistic Lock 적용)
	if existingVital.Version != request.Version {
		return pkgError.WrapWithCode(pkgError.EmptyBusinessError(), pkgError.Conflict, "version mismatch")
	}

	existingVital.Value = request.Value
	existingVital.Version = request.Version + 1
	existingVital.UpdatedAt = &now

	if err := v.repo.UpdateVital(ctx, existingVital); err != nil {
		return pkgError.Wrap(err)
	}

	return nil
}

func NewVitalService(repo vital.VitalRepository) vital.VitalService {
	return &vitalService{repo}
}
