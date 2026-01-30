package controller

import "aitrics-vital-signs/api-server/domain/patient"

type patientController struct {
	service patient.PatientService
}

func NewPatientController(service patient.PatientService) patient.PatientController {
	p := &patientController{
		service: service,
	}

	return p
}
