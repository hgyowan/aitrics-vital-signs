package controller

import (
	"aitrics-vital-signs/api-server/domain/mock"
	"aitrics-vital-signs/api-server/domain/vital"
	pkgError "aitrics-vital-signs/library/error"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var (
	testVitalController vital.VitalController
	mockVitalService    *mock.MockVitalService
)

func beforeEachVital(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockVitalService = mock.NewMockVitalService(ctrl)
	testVitalController = NewVitalController(mockVitalService)
}

func Test_UpsertVital(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		body           string
		mockSetup      func(svc *mock.MockVitalService)
		wantStatusCode int
	}{
		{
			name: "성공 - INSERT (새 데이터)",
			body: `{
				"patient_id": "P00001234",
				"recorded_at": "2025-12-01T10:15:00Z",
				"vital_type": "HR",
				"value": 110.0,
				"version": 1
			}`,
			mockSetup: func(svc *mock.MockVitalService) {
				svc.EXPECT().
					UpsertVital(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "성공 - UPDATE (기존 데이터)",
			body: `{
				"patient_id": "P00001234",
				"recorded_at": "2025-12-01T10:15:00Z",
				"vital_type": "HR",
				"value": 115.0,
				"version": 1
			}`,
			mockSetup: func(svc *mock.MockVitalService) {
				svc.EXPECT().
					UpsertVital(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "실패 - 필수 필드 누락 (patient_id 없음)",
			body: `{
				"recorded_at": "2025-12-01T10:15:00Z",
				"vital_type": "HR",
				"value": 110.0,
				"version": 1
			}`,
			mockSetup:      func(svc *mock.MockVitalService) {},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "실패 - 필수 필드 누락 (version 없음)",
			body: `{
				"patient_id": "P00001234",
				"recorded_at": "2025-12-01T10:15:00Z",
				"vital_type": "HR",
				"value": 110.0
			}`,
			mockSetup:      func(svc *mock.MockVitalService) {},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "실패 - 잘못된 VitalType",
			body: `{
				"patient_id": "P00001234",
				"recorded_at": "2025-12-01T10:15:00Z",
				"vital_type": "INVALID",
				"value": 110.0,
				"version": 1
			}`,
			mockSetup:      func(svc *mock.MockVitalService) {},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "실패 - 잘못된 날짜 형식",
			body: `{
				"patient_id": "P00001234",
				"recorded_at": "2025-12-01 10:15:00",
				"vital_type": "HR",
				"value": 110.0,
				"version": 1
			}`,
			mockSetup:      func(svc *mock.MockVitalService) {},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "실패 - Version Conflict (Optimistic Lock)",
			body: `{
				"patient_id": "P00001234",
				"recorded_at": "2025-12-01T10:15:00Z",
				"vital_type": "HR",
				"value": 120.0,
				"version": 1
			}`,
			mockSetup: func(svc *mock.MockVitalService) {
				svc.EXPECT().
					UpsertVital(gomock.Any(), gomock.Any()).
					Return(pkgError.WrapWithCode(pkgError.EmptyBusinessError(), pkgError.Conflict, "version mismatch"))
			},
			wantStatusCode: http.StatusConflict,
		},
		{
			name: "실패 - INSERT 시 잘못된 version",
			body: `{
				"patient_id": "P00001234",
				"recorded_at": "2025-12-01T10:15:00Z",
				"vital_type": "HR",
				"value": 110.0,
				"version": 5
			}`,
			mockSetup: func(svc *mock.MockVitalService) {
				svc.EXPECT().
					UpsertVital(gomock.Any(), gomock.Any()).
					Return(pkgError.WrapWithCode(pkgError.EmptyBusinessError(), pkgError.WrongParam, "version must be 1 for new record"))
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "실패 - 비즈니스 로직 에러 (500)",
			body: `{
				"patient_id": "P00001234",
				"recorded_at": "2025-12-01T10:15:00Z",
				"vital_type": "HR",
				"value": 110.0,
				"version": 1
			}`,
			mockSetup: func(svc *mock.MockVitalService) {
				svc.EXPECT().
					UpsertVital(gomock.Any(), gomock.Any()).
					Return(pkgError.WrapWithCode(pkgError.EmptyBusinessError(), pkgError.Create))
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeEachVital(t)
			tt.mockSetup(mockVitalService)

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest(
				http.MethodPost,
				"/api/v1/vitals",
				strings.NewReader(tt.body),
			)
			req.Header.Set("Content-Type", "application/json")
			ctx.Request = req

			testVitalController.UpsertVital(ctx)

			require.Equal(t, tt.wantStatusCode, w.Code)
		})
	}
}
