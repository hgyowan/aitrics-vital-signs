package service

import (
	"aitrics-vital-signs/api-server/domain/inference"
	"aitrics-vital-signs/api-server/domain/vital"
	"aitrics-vital-signs/api-server/pkg/constant"
	"aitrics-vital-signs/library/envs"
	pkgError "aitrics-vital-signs/library/error"
	"context"
	"time"
)

type inferenceService struct {
	vitalRepo vital.VitalRepository
}

func (i *inferenceService) CalculateVitalRisk(ctx context.Context, request inference.VitalRiskRequest) (*inference.VitalRiskResponse, error) {
	// 환경변수에서 시간 범위 읽기 (기본값: 24시간)
	timeWindowHours := envs.VitalRiskTimeWindowHours

	// 현재 시간 기준으로 시간 범위 설정
	now := time.Now().UTC()
	from := now.Add(-time.Duration(timeWindowHours) * time.Hour)
	to := now

	// Vital 데이터 조회 (HR, SBP, SpO2만 조회)
	vitals, err := i.vitalRepo.FindVitalsByPatientIDAndDateRange(ctx, request.PatientID, from, to, "")
	if err != nil {
		return nil, pkgError.WrapWithCode(err, pkgError.Get)
	}

	// 각 Vital Type별로 데이터 수집
	vitalData := make(map[string][]float64)
	for _, v := range vitals {
		// HR, SBP, SpO2만 처리
		if v.VitalType == constant.VitalTypeHR.String() || v.VitalType == constant.VitalTypeSBP.String() || v.VitalType == constant.VitalTypeSpO2.String() {
			vitalData[v.VitalType] = append(vitalData[v.VitalType], v.Value)
		}
	}

	// 각 Vital Type별 평균 계산
	vitalAverages := make(map[string]float64)
	for vitalType, values := range vitalData {
		if len(values) > 0 {
			sum := 0.0
			for _, val := range values {
				sum += val
			}
			vitalAverages[vitalType] = sum / float64(len(values))
		}
	}

	// 위험 조건 평가
	var triggeredRules []string

	// HR > 120
	if avg, exists := vitalAverages[constant.VitalTypeHR.String()]; exists && avg > 120 {
		triggeredRules = append(triggeredRules, "HR > 120")
	}

	// SBP < 90
	if avg, exists := vitalAverages[constant.VitalTypeSBP.String()]; exists && avg < 90 {
		triggeredRules = append(triggeredRules, "SBP < 90")
	}

	// SpO2 < 90
	if avg, exists := vitalAverages[constant.VitalTypeSpO2.String()]; exists && avg < 90 {
		triggeredRules = append(triggeredRules, "SpO2 < 90")
	}

	// risk_level 결정
	riskLevel := "LOW"
	triggeredCount := len(triggeredRules)
	if triggeredCount >= 3 {
		riskLevel = "HIGH"
	} else if triggeredCount >= 1 {
		riskLevel = "MEDIUM"
	}

	return &inference.VitalRiskResponse{
		PatientID:          request.PatientID,
		RiskLevel:          riskLevel,
		TriggeredRules:     triggeredRules,
		VitalAverages:      vitalAverages,
		DataPointsAnalyzed: len(vitals),
		TimeRange: inference.TimeRange{
			From: from,
			To:   to,
		},
		EvaluatedAt: now,
	}, nil
}

func NewInferenceService(vitalRepo vital.VitalRepository) inference.InferenceService {
	return &inferenceService{
		vitalRepo: vitalRepo,
	}
}
