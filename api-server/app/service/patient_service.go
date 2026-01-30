package service

import (
	"aitrics-vital-signs/api-server/domain/patient"
	pkgError "aitrics-vital-signs/library/error"
	"context"
	"time"

	"github.com/google/uuid"
)

type patientService struct {
	repo patient.PatientRepository
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
		return pkgError.WrapWithCode(err, pkgError.Create)
	}

	return nil
}

func NewPatientService(repo patient.PatientRepository) patient.PatientService {
	return &patientService{repo}
}
