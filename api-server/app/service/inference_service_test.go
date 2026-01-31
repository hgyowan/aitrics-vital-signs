package service

import (
	"aitrics-vital-signs/api-server/domain/inference"
	"aitrics-vital-signs/api-server/domain/mock"
	"aitrics-vital-signs/api-server/domain/vital"
	"aitrics-vital-signs/api-server/pkg/constant"
	"aitrics-vital-signs/library/envs"
	pkgError "aitrics-vital-signs/library/error"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var (
	mockVitalSvc *mock.MockVitalService
	inferenceSvc     inference.InferenceService
)

func beforeEachInference(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockVitalSvc = mock.NewMockVitalService(ctrl)
	inferenceSvc = NewInferenceService(mockVitalSvc)
}

func Test_CalculateVitalRisk(t *testing.T) {
	// envs.VitalRiskTimeWindowHours 기본값(24시간) 사용
	// 환경변수는 패키지 초기화 시점에 로드되므로 테스트 실행 전 설정 필요
	now := time.Now().UTC()

	tests := []struct {
		name               string
		req                inference.VitalRiskRequest
		setupMock          func()
		wantErr            bool
		expectedErr        error
		expectedRiskLevel  string
		expectedRulesCount int
	}{
		{
			name: "성공 - HIGH 위험 (모든 조건 충족)",
			req: inference.VitalRiskRequest{
				PatientID: "P00001234",
			},
			setupMock: func() {
				response := &vital.GetVitalsResponse{
					PatientID: "P00001234",
					Items: []vital.VitalItemResponse{
						// HR > 120 (평균: 130)
						{RecordedAt: now.Add(-1 * time.Hour), VitalType: constant.VitalTypeHR.String(), Value: 125.0},
						{RecordedAt: now.Add(-2 * time.Hour), VitalType: constant.VitalTypeHR.String(), Value: 135.0},
						// SBP < 90 (평균: 85)
						{RecordedAt: now.Add(-1 * time.Hour), VitalType: constant.VitalTypeSBP.String(), Value: 82.0},
						{RecordedAt: now.Add(-2 * time.Hour), VitalType: constant.VitalTypeSBP.String(), Value: 88.0},
						// SpO2 < 90 (평균: 87)
						{RecordedAt: now.Add(-1 * time.Hour), VitalType: constant.VitalTypeSpO2.String(), Value: 85.0},
						{RecordedAt: now.Add(-2 * time.Hour), VitalType: constant.VitalTypeSpO2.String(), Value: 89.0},
					},
				}
				mockVitalSvc.EXPECT().
					GetVitalsByPatientIDAndDateRange(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, req vital.GetVitalsRequest) (*vital.GetVitalsResponse, error) {
						// 시간 범위 검증
						require.True(t, req.From.Before(req.To))
						require.True(t, req.To.Sub(req.From) == 24*time.Hour)
						return response, nil
					})
			},
			wantErr:            false,
			expectedRiskLevel:  "HIGH",
			expectedRulesCount: 3,
		},
		{
			name: "성공 - MEDIUM 위험 (2개 조건 충족)",
			req: inference.VitalRiskRequest{
				PatientID: "P00001234",
			},
			setupMock: func() {
				response := &vital.GetVitalsResponse{
					PatientID: "P00001234",
					Items: []vital.VitalItemResponse{
						// HR > 120 (평균: 130)
						{RecordedAt: now.Add(-1 * time.Hour), VitalType: constant.VitalTypeHR.String(), Value: 125.0},
						{RecordedAt: now.Add(-2 * time.Hour), VitalType: constant.VitalTypeHR.String(), Value: 135.0},
						// SBP 정상 (평균: 115)
						{RecordedAt: now.Add(-1 * time.Hour), VitalType: constant.VitalTypeSBP.String(), Value: 110.0},
						{RecordedAt: now.Add(-2 * time.Hour), VitalType: constant.VitalTypeSBP.String(), Value: 120.0},
						// SpO2 < 90 (평균: 87)
						{RecordedAt: now.Add(-1 * time.Hour), VitalType: constant.VitalTypeSpO2.String(), Value: 85.0},
						{RecordedAt: now.Add(-2 * time.Hour), VitalType: constant.VitalTypeSpO2.String(), Value: 89.0},
					},
				}
				mockVitalSvc.EXPECT().
					GetVitalsByPatientIDAndDateRange(gomock.Any(), gomock.Any()).
					Return(response, nil)
			},
			wantErr:            false,
			expectedRiskLevel:  "MEDIUM",
			expectedRulesCount: 2,
		},
		{
			name: "성공 - MEDIUM 위험 (1개 조건 충족)",
			req: inference.VitalRiskRequest{
				PatientID: "P00001234",
			},
			setupMock: func() {
				response := &vital.GetVitalsResponse{
					PatientID: "P00001234",
					Items: []vital.VitalItemResponse{
						// HR > 120 (평균: 130)
						{RecordedAt: now.Add(-1 * time.Hour), VitalType: constant.VitalTypeHR.String(), Value: 125.0},
						{RecordedAt: now.Add(-2 * time.Hour), VitalType: constant.VitalTypeHR.String(), Value: 135.0},
						// SBP 정상 (평균: 115)
						{RecordedAt: now.Add(-1 * time.Hour), VitalType: constant.VitalTypeSBP.String(), Value: 110.0},
						{RecordedAt: now.Add(-2 * time.Hour), VitalType: constant.VitalTypeSBP.String(), Value: 120.0},
						// SpO2 정상 (평균: 97)
						{RecordedAt: now.Add(-1 * time.Hour), VitalType: constant.VitalTypeSpO2.String(), Value: 95.0},
						{RecordedAt: now.Add(-2 * time.Hour), VitalType: constant.VitalTypeSpO2.String(), Value: 99.0},
					},
				}
				mockVitalSvc.EXPECT().
					GetVitalsByPatientIDAndDateRange(gomock.Any(), gomock.Any()).
					Return(response, nil)
			},
			wantErr:            false,
			expectedRiskLevel:  "MEDIUM",
			expectedRulesCount: 1,
		},
		{
			name: "성공 - LOW 위험 (조건 충족 없음)",
			req: inference.VitalRiskRequest{
				PatientID: "P00001234",
			},
			setupMock: func() {
				response := &vital.GetVitalsResponse{
					PatientID: "P00001234",
					Items: []vital.VitalItemResponse{
						// HR 정상 (평균: 80)
						{RecordedAt: now.Add(-1 * time.Hour), VitalType: constant.VitalTypeHR.String(), Value: 75.0},
						{RecordedAt: now.Add(-2 * time.Hour), VitalType: constant.VitalTypeHR.String(), Value: 85.0},
						// SBP 정상 (평균: 115)
						{RecordedAt: now.Add(-1 * time.Hour), VitalType: constant.VitalTypeSBP.String(), Value: 110.0},
						{RecordedAt: now.Add(-2 * time.Hour), VitalType: constant.VitalTypeSBP.String(), Value: 120.0},
						// SpO2 정상 (평균: 97)
						{RecordedAt: now.Add(-1 * time.Hour), VitalType: constant.VitalTypeSpO2.String(), Value: 95.0},
						{RecordedAt: now.Add(-2 * time.Hour), VitalType: constant.VitalTypeSpO2.String(), Value: 99.0},
					},
				}
				mockVitalSvc.EXPECT().
					GetVitalsByPatientIDAndDateRange(gomock.Any(), gomock.Any()).
					Return(response, nil)
			},
			wantErr:            false,
			expectedRiskLevel:  "LOW",
			expectedRulesCount: 0,
		},
		{
			name: "성공 - 데이터 없음 (LOW)",
			req: inference.VitalRiskRequest{
				PatientID: "P99999999",
			},
			setupMock: func() {
				response := &vital.GetVitalsResponse{
					PatientID: "P99999999",
					Items:     []vital.VitalItemResponse{},
				}
				mockVitalSvc.EXPECT().
					GetVitalsByPatientIDAndDateRange(gomock.Any(), gomock.Any()).
					Return(response, nil)
			},
			wantErr:            false,
			expectedRiskLevel:  "LOW",
			expectedRulesCount: 0,
		},
		{
			name: "실패 - Repository 에러",
			req: inference.VitalRiskRequest{
				PatientID: "P00001234",
			},
			setupMock: func() {
				mockVitalSvc.EXPECT().
					GetVitalsByPatientIDAndDateRange(gomock.Any(), gomock.Any()).
					Return(nil, pkgError.WrapWithCode(pkgError.EmptyBusinessError(), pkgError.Get, "db error"))
			},
			wantErr:     true,
			expectedErr: pkgError.WrapWithCode(pkgError.EmptyBusinessError(), pkgError.Get),
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeEachInference(t)
			tt.setupMock()

			result, err := inferenceSvc.CalculateVitalRisk(ctx, tt.req)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, result)
				if tt.expectedErr != nil {
					expectedBE, _ := pkgError.CastBusinessError(tt.expectedErr)
					actualBE, _ := pkgError.CastBusinessError(err)
					require.Equal(t, expectedBE.Status.Code, actualBE.Status.Code)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Equal(t, tt.req.PatientID, result.PatientID)
				require.Equal(t, tt.expectedRiskLevel, result.RiskLevel)
				require.Equal(t, tt.expectedRulesCount, len(result.TriggeredRules))
				require.NotNil(t, result.TimeRange)
				require.NotNil(t, result.EvaluatedAt)
			}
		})
	}
}

func Test_CalculateVitalRisk_EnvironmentVariable(t *testing.T) {
	// 환경변수 테스트
	// envs.VitalRiskTimeWindowHours는 패키지 초기화 시점에 로드됨
	// 테스트 실행 전 VITAL_RISK_TIME_WINDOW_HOURS 환경변수를 설정해야 함
	// 기본값은 24시간
	beforeEachInference(t)

	timeWindow := envs.VitalRiskTimeWindowHours
	mockVitalSvc.EXPECT().
		GetVitalsByPatientIDAndDateRange(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, req vital.GetVitalsRequest) (*vital.GetVitalsResponse, error) {
			// 환경변수에 설정된 시간 범위인지 확인
			require.True(t, req.To.Sub(req.From) == time.Duration(timeWindow)*time.Hour)
			return &vital.GetVitalsResponse{
				PatientID: "P00001234",
				Items:     []vital.VitalItemResponse{},
			}, nil
		})

	req := inference.VitalRiskRequest{PatientID: "P00001234"}
	result, err := inferenceSvc.CalculateVitalRisk(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, result)
}
