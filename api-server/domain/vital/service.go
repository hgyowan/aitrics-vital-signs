//go:generate mockgen -source=service.go -destination=../mock/mock_vital_service.go -package=mock
package vital

import "context"

type VitalService interface {
	UpsertVital(ctx context.Context, request UpsertVitalRequest) error
}