//go:generate mockgen -source=repository.go -destination=../mock/mock_vital_repository.go -package=mock
package vital

import (
	"context"
)

type VitalRepository interface {
	FindVitalByPatientIDAndRecordedAtAndVitalType(ctx context.Context, param FindVitalByPatientIDAndRecordedAtAndVitalTypeParam) (*Vital, error)
	FindVitalsByPatientIDAndDateRange(ctx context.Context, param FindVitalsByPatientIDAndDateRangeParam) ([]Vital, error)
	CreateVital(ctx context.Context, model *Vital) error
	UpdateVital(ctx context.Context, model *Vital) error
}
