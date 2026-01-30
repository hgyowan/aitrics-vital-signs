//go:generate mockgen -source=repository.go -destination=../mock/mock_repository.go -package=mock
package patient

import "context"

type PatientRepository interface {
	CreatePatient(ctx context.Context, model *Patient) error
}
