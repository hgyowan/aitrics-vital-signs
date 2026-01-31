package inference

import "time"

type VitalRiskRequest struct {
	PatientID string `json:"patient_id" binding:"required"`
}

type VitalRiskResponse struct {
	PatientID          string            `json:"patient_id"`
	RiskLevel          string            `json:"risk_level"`
	TriggeredRules     []string          `json:"triggered_rules"`
	VitalAverages      map[string]float64 `json:"vital_averages"`
	DataPointsAnalyzed int               `json:"data_points_analyzed"`
	TimeRange          TimeRange         `json:"time_range"`
	EvaluatedAt        time.Time         `json:"evaluated_at"`
}

type TimeRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}