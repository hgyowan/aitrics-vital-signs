//go:generate mockgen -source=service.go -destination=../mock/mock_patient_service.go -package=mock
package patient

import (
	"aitrics-vital-signs/api-server/domain/vital"
	"context"
)

type PatientService interface {
	CreatePatient(ctx context.Context, request CreatePatientRequest) error
	UpdatePatient(ctx context.Context, patientID string, request UpdatePatientRequest) error
	GetPatientVitals(ctx context.Context, patientID string, request GetPatientVitalsQueryRequest) (*vital.GetVitalsResponse, error)
}
