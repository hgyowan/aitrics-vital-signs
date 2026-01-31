package vital

import "time"

type UpsertVitalRequest struct {
	PatientID  string    `json:"patient_id" binding:"required"`
	RecordedAt time.Time `json:"recorded_at" binding:"required"`
	VitalType  string    `json:"vital_type" binding:"required,oneof=HR RR SBP DBP SpO2 BT"`
	Value      float64   `json:"value" binding:"required"`
	Version    int       `json:"version" binding:"required,min=1"`
}

type GetVitalsRequest struct {
	PatientID string
	From      time.Time
	To        time.Time
	VitalType string // optional: 있으면 해당 타입만, 없으면 모든 타입
}

// vital_type 유무와 관계없이 동일한 응답 구조
type GetVitalsResponse struct {
	PatientID string              `json:"patient_id"`
	Items     []VitalItemResponse `json:"items"`
}

type VitalItemResponse struct {
	VitalType  string    `json:"vital_type"`
	RecordedAt time.Time `json:"recorded_at"`
	Value      float64   `json:"value"`
}