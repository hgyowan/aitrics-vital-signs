package service

import (
	"aitrics-vital-signs/api-server/domain/patient"
	"aitrics-vital-signs/api-server/domain/vital"
	pkgError "aitrics-vital-signs/library/error"
	"context"
	"time"

	"github.com/google/uuid"
)

type patientService struct {
	repo         patient.PatientRepository
	vitalService vital.VitalService
}

func (p *patientService) CreatePatient(ctx context.Context, request patient.CreatePatientRequest) error {
	birthDate, err := time.Parse(time.DateOnly, request.BirthDate)
	if err != nil {
		return pkgError.WrapWithCode(err, pkgError.WrongParam)
	}

	now := time.Now().UTC()
	if err := p.repo.CreatePatient(ctx, &patient.Patient{
		ID:        uuid.NewString(),
		PatientID: request.PatientID,
		Name:      request.Name,
		Gender:    request.Gender,
		BirthDate: birthDate,
		CreatedAt: now,
		UpdatedAt: &now,
	}); err != nil {
		return pkgError.Wrap(err)
	}

	return nil
}

func (p *patientService) UpdatePatient(ctx context.Context, patientID string, request patient.UpdatePatientRequest) error {
	birthDate, err := time.Parse(time.DateOnly, request.BirthDate)
	if err != nil {
		return pkgError.WrapWithCode(err, pkgError.WrongParam)
	}

	existingPatient, err := p.repo.FindPatientByID(ctx, patientID)
	if err != nil {
		return pkgError.Wrap(err)
	}

	if existingPatient.Version != request.Version {
		return pkgError.WrapWithCode(pkgError.EmptyBusinessError(), pkgError.Conflict, "version mismatch")
	}

	now := time.Now().UTC()
	existingPatient.Name = request.Name
	existingPatient.Gender = request.Gender
	existingPatient.BirthDate = birthDate
	existingPatient.Version = request.Version + 1
	existingPatient.UpdatedAt = &now

	if err := p.repo.UpdatePatient(ctx, existingPatient); err != nil {
		return pkgError.Wrap(err)
	}

	return nil
}

func (p *patientService) GetPatientVitals(ctx context.Context, patientID string, request patient.GetPatientVitalsQueryRequest) (*vital.GetVitalsResponse, error) {
	// Query Parameter 날짜 파싱
	from, err := time.Parse(time.RFC3339, request.From)
	if err != nil {
		return nil, pkgError.WrapWithCode(err, pkgError.WrongParam, "invalid from date format")
	}

	to, err := time.Parse(time.RFC3339, request.To)
	if err != nil {
		return nil, pkgError.WrapWithCode(err, pkgError.WrongParam, "invalid to date format")
	}

	// Vital Service를 통해 데이터 조회
	return p.vitalService.GetVitalsByPatientIDAndDateRange(ctx, vital.GetVitalsRequest{
		PatientID: patientID,
		From:      from,
		To:        to,
		VitalType: request.VitalType,
	})
}

func NewPatientService(repo patient.PatientRepository, vitalService vital.VitalService) patient.PatientService {
	return &patientService{
		repo:         repo,
		vitalService: vitalService,
	}
}
