package repository

import (
	"aitrics-vital-signs/api-server/domain"
	"aitrics-vital-signs/api-server/domain/patient"
)

type patientRepository struct {
	externalGormClient domain.ExternalDBClient
}

func NewPatientRepository(externalGormClient domain.ExternalDBClient) patient.PatientRepository {
	return &patientRepository{externalGormClient: externalGormClient}
}
