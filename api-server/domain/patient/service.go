//go:generate mockgen -source=service.go -destination=../mock/mock_service.go -package=mock
package patient

type PatientService interface {
}
