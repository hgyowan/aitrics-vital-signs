package service

import "aitrics-vital-signs/api-server/domain/patient"

type patientService struct {
	repo patient.PatientRepository
}

func NewPatientService(repo patient.PatientRepository) patient.PatientService {
	return &patientService{repo}
}
