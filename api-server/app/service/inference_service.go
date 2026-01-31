package service

import (
	"aitrics-vital-signs/api-server/domain/inference"
	"aitrics-vital-signs/api-server/domain/vital"
	"aitrics-vital-signs/api-server/pkg/constant"
	"aitrics-vital-signs/library/envs"
	pkgError "aitrics-vital-signs/library/error"
	"context"
	"fmt"
	"math"
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

	// HR, SBP, SpO2 만 처리
	vitals, err := i.vitalRepo.FindVitalsByPatientIDAndDateRange(ctx, vital.FindVitalsByPatientIDAndDateRangeParam{
		PatientID:  request.PatientID,
		From:       from,
		To:         to,
		VitalTypes: []string{constant.VitalTypeHR.String(), constant.VitalTypeSBP.String(), constant.VitalTypeSpO2.String()},
	})
	if err != nil {
		return nil, pkgError.Wrap(err)
	}

	// 각 Vital Type별로 데이터 수집
	vitalData := make(map[string][]float64)
	for _, v := range vitals {
		vitalData[v.VitalType] = append(vitalData[v.VitalType], v.Value)
	}

	// 각 Vital Type별 평균 계산
	vitalAverages := make(map[string]float64)
	for vitalType, values := range vitalData {
		if len(values) > 0 {
			sum := 0.0
			for _, val := range values {
				sum += val
			}
			vitalAverages[vitalType] = math.Round((sum/float64(len(values)))*10) / 10
		}
	}

	// 위험 조건 평가
	var triggeredRules []string

	// HR > 120
	if avg, exists := vitalAverages[constant.VitalTypeHR.String()]; exists && avg > 120 {
		triggeredRules = append(triggeredRules, fmt.Sprintf("%s > 120", constant.VitalTypeHR.String()))
	}

	// SBP < 90
	if avg, exists := vitalAverages[constant.VitalTypeSBP.String()]; exists && avg < 90 {
		triggeredRules = append(triggeredRules, fmt.Sprintf("%s < 90", constant.VitalTypeSBP.String()))
	}

	// SpO2 < 90
	if avg, exists := vitalAverages[constant.VitalTypeSpO2.String()]; exists && avg < 90 {
		triggeredRules = append(triggeredRules, fmt.Sprintf("%s < 90", constant.VitalTypeSpO2.String()))
	}

	// risk_level 결정
	riskLevel := constant.RiskLevelLow
	triggeredCount := len(triggeredRules)
	if triggeredCount >= 3 {
		riskLevel = constant.RiskLevelHigh
	} else if triggeredCount >= 1 {
		riskLevel = constant.RiskLevelMedium
	}

	return &inference.VitalRiskResponse{
		PatientID:          request.PatientID,
		RiskLevel:          riskLevel.String(),
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
