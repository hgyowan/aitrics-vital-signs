//go:generate mockgen -source=controller.go -destination=../mock/mock_controller.go -package=mock
package patient

type PatientController interface {
}
