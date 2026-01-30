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
	return pkgError.Wrap(p.externalGormClient.MySQL().WithContext(ctx).Create(model).Error)
}

func NewPatientRepository(externalGormClient domain.ExternalDBClient) patient.PatientRepository {
	return &patientRepository{externalGormClient: externalGormClient}
}
