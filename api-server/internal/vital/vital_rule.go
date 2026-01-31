package vital

import "aitrics-vital-signs/api-server/pkg/constant"

type Comparator string

const (
	CompGT  Comparator = ">"
	CompGTE Comparator = ">="
	CompLT  Comparator = "<"
	CompLTE Comparator = "<="
)

type RiskRule struct {
	VitalType  string
	Comparator Comparator
	Threshold  float64
}

var RiskRules = []RiskRule{
	{
		VitalType:  constant.VitalTypeHR.String(),
		Comparator: CompGT,
		Threshold:  120,
	},
	{
		VitalType:  constant.VitalTypeSBP.String(),
		Comparator: CompLT,
		Threshold:  90,
	},
	{
		VitalType:  constant.VitalTypeSpO2.String(),
		Comparator: CompLT,
		Threshold:  90,
	},
}

func EvaluateRule(value float64, rule RiskRule) bool {
	switch rule.Comparator {
	case CompGT:
		return value > rule.Threshold
	case CompGTE:
		return value >= rule.Threshold
	case CompLT:
		return value < rule.Threshold
	case CompLTE:
		return value <= rule.Threshold
	default:
		return false
	}
}
