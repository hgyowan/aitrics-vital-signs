package constant

// Vital Type 상수 정의
type VitalType string

const (
	VitalTypeHR   VitalType = "HR"   // Heart Rate (심박수)
	VitalTypeSBP  VitalType = "SBP"  // Systolic Blood Pressure (수축기 혈압)
	VitalTypeDBP  VitalType = "DBP"  // Diastolic Blood Pressure (이완기 혈압)
	VitalTypeSpO2 VitalType = "SpO2" // Oxygen Saturation (산소포화도)
	VitalTypeRR   VitalType = "RR"   // Respiratory Rate (호흡수)
	VitalTypeBT   VitalType = "BT"   // Body Temperature (체온)
)

func (v VitalType) String() string {
	return string(v)
}

type RiskLevel string

const (
	RiskLevelLow    RiskLevel = "LOW"
	RiskLevelMedium RiskLevel = "MEDIUM"
	RiskLevelHigh   RiskLevel = "HIGH"
)

func (r RiskLevel) String() string {
	return string(r)
}
