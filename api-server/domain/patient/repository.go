//go:generate mockgen -source=repository.go -destination=../mock/mock_patient_repository.go -package=mock
package patient

import "context"

type PatientRepository interface {
	CreatePatient(ctx context.Context, model *Patient) error
	FindPatientByID(ctx context.Context, patientID string) (*Patient, error)
	UpdatePatient(ctx context.Context, model *Patient) error
}
