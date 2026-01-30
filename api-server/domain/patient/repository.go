//go:generate mockgen -source=repository.go -destination=../mock/mock_repository.go -package=mock
package patient

type PatientRepository interface {
}
