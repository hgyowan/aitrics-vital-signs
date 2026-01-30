//go:generate mockgen -source=repository.go -destination=../mock/mock_vital_repository.go -package=mock
package vital

import (
	"context"
	"time"
)

type VitalRepository interface {
	FindVitalByPatientIDAndRecordedAtAndVitalType(ctx context.Context, patientID string, recordedAt time.Time, vitalType string) (*Vital, error)
	CreateVital(ctx context.Context, model *Vital) error
	UpdateVital(ctx context.Context, model *Vital) error
}