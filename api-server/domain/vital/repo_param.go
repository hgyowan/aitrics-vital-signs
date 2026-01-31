package vital

import "time"

type FindVitalByPatientIDAndRecordedAtAndVitalTypeParam struct {
	PatientID  string
	RecordedAt time.Time
	VitalType  string
}

type FindVitalsByPatientIDAndDateRangeParam struct {
	PatientID string
	From      time.Time
	To        time.Time
	VitalType string
}
