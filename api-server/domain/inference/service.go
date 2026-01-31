//go:generate mockgen -source=service.go -destination=../mock/mock_inference_service.go -package=mock
package inference

import "context"

type InferenceService interface {
	CalculateVitalRisk(ctx context.Context, request VitalRiskRequest) (*VitalRiskResponse, error)
}