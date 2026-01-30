//go:generate mockgen -source=service.go -destination=../mock/mock_service.go -package=mock
package patient

import "context"

type PatientService interface {
	CreatePatient(ctx context.Context, request CreatePatientRequest) error
	UpdatePatient(ctx context.Context, patientID string, request UpdatePatientRequest) error
}
