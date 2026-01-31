package patient

import "time"

type CreatePatientRequest struct {
	PatientID string `json:"patientId" binding:"required"`
	Name      string `json:"name" binding:"required"`
	Gender    string `json:"gender" binding:"required,oneof=M F"`
	BirthDate string `json:"birthDate" binding:"required,datetime=2006-01-02"`
}

type UpdatePatientRequest struct {
	Name      string `json:"name" binding:"required"`
	Gender    string `json:"gender" binding:"required,oneof=M F"`
	BirthDate string `json:"birthDate" binding:"required,datetime=2006-01-02"`
	Version   int    `json:"version" binding:"required,min=1"`
}

type GetPatientVitalsRequest struct {
	From      string `form:"from" binding:"required"` // RFC3339 format
	To        string `form:"to" binding:"required"`   // RFC3339 format
	VitalType string `form:"vital_type" binding:"omitempty,oneof=HR RR SBP DBP SpO2 BT"`
}

type GetPatientVitalsResponse struct {
	PatientID string              `json:"patient_id"`
	Items     []VitalItemResponse `json:"items"`
}

type VitalItemResponse struct {
	VitalType  string    `json:"vital_type"`
	RecordedAt time.Time `json:"recorded_at"`
	Value      float64   `json:"value"`
}
