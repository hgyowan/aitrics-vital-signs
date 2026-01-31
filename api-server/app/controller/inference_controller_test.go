package controller

import (
	"aitrics-vital-signs/api-server/domain/inference"
	"aitrics-vital-signs/api-server/domain/mock"
	pkgError "aitrics-vital-signs/library/error"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var (
	testInferenceController inference.InferenceController
	mockInferenceService    *mock.MockInferenceService
)

func beforeEachInference(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockInferenceService = mock.NewMockInferenceService(ctrl)
	testInferenceController = NewInferenceController(mockInferenceService)
}

func Test_CalculateVitalRisk(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		body           string
		mockSetup      func(svc *mock.MockInferenceService)
		wantStatusCode int
	}{
		{
			name: "성공 - HIGH 위험",
			body: `{
				"patient_id": "P00001234"
			}`,
			mockSetup: func(svc *mock.MockInferenceService) {
				svc.EXPECT().
					CalculateVitalRisk(gomock.Any(), gomock.Any()).
					Return(&inference.VitalRiskResponse{
						PatientID: "P00001234",
						RiskLevel: "HIGH",
						TriggeredRules: []string{
							"HR > 120",
							"SBP < 90",
							"SpO2 < 90",
						},
						VitalAverages: map[string]float64{
							"HR":   135.2,
							"SBP":  82.5,
							"SpO2": 87.3,
						},
						DataPointsAnalyzed: 48,
						TimeRange: inference.TimeRange{
							From: time.Now().Add(-24 * time.Hour),
							To:   time.Now(),
						},
						EvaluatedAt: time.Now(),
					}, nil)
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "성공 - MEDIUM 위험",
			body: `{
				"patient_id": "P00001234"
			}`,
			mockSetup: func(svc *mock.MockInferenceService) {
				svc.EXPECT().
					CalculateVitalRisk(gomock.Any(), gomock.Any()).
					Return(&inference.VitalRiskResponse{
						PatientID: "P00001234",
						RiskLevel: "MEDIUM",
						TriggeredRules: []string{
							"HR > 120",
						},
						VitalAverages: map[string]float64{
							"HR":   130.5,
							"SBP":  110.0,
							"SpO2": 95.0,
						},
						DataPointsAnalyzed: 24,
						TimeRange: inference.TimeRange{
							From: time.Now().Add(-24 * time.Hour),
							To:   time.Now(),
						},
						EvaluatedAt: time.Now(),
					}, nil)
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "성공 - LOW 위험",
			body: `{
				"patient_id": "P00001234"
			}`,
			mockSetup: func(svc *mock.MockInferenceService) {
				svc.EXPECT().
					CalculateVitalRisk(gomock.Any(), gomock.Any()).
					Return(&inference.VitalRiskResponse{
						PatientID:          "P00001234",
						RiskLevel:          "LOW",
						TriggeredRules:     []string{},
						VitalAverages: map[string]float64{
							"HR":   80.0,
							"SBP":  115.0,
							"SpO2": 97.0,
						},
						DataPointsAnalyzed: 24,
						TimeRange: inference.TimeRange{
							From: time.Now().Add(-24 * time.Hour),
							To:   time.Now(),
						},
						EvaluatedAt: time.Now(),
					}, nil)
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "실패 - patient_id 필드 없음",
			body: `{}`,
			mockSetup:      func(svc *mock.MockInferenceService) {},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "실패 - 잘못된 JSON",
			body: `{
				"patient_id": "P00001234",
			}`,
			mockSetup:      func(svc *mock.MockInferenceService) {},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "실패 - Service 에러 (DB 조회 실패)",
			body: `{
				"patient_id": "P00001234"
			}`,
			mockSetup: func(svc *mock.MockInferenceService) {
				svc.EXPECT().
					CalculateVitalRisk(gomock.Any(), gomock.Any()).
					Return(nil, pkgError.WrapWithCode(pkgError.EmptyBusinessError(), pkgError.Get, "db error"))
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeEachInference(t)
			tt.mockSetup(mockInferenceService)

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest(
				http.MethodPost,
				"/api/v1/inference/vital-risk",
				strings.NewReader(tt.body),
			)
			req.Header.Set("Content-Type", "application/json")
			ctx.Request = req

			testInferenceController.CalculateVitalRisk(ctx)

			require.Equal(t, tt.wantStatusCode, w.Code)
		})
	}
}
