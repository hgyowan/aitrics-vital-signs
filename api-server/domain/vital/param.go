package vital

import "time"

type UpsertVitalRequest struct {
	PatientID  string    `json:"patient_id" binding:"required"`
	RecordedAt time.Time `json:"recorded_at" binding:"required"`
	VitalType  string    `json:"vital_type" binding:"required,oneof=HR RR SBP DBP SpO2 BT"`
	Value      float64   `json:"value" binding:"required"`
	Version    int       `json:"version" binding:"required,min=1"`
}