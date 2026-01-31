package controller

import (
	"aitrics-vital-signs/api-server/domain/mock"
	"aitrics-vital-signs/api-server/domain/patient"
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
	controller  patient.PatientController
	mockService *mock.MockPatientService
)

func beforeEach(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockService = mock.NewMockPatientService(ctrl)
	controller = NewPatientController(mockService)
}

func Test_CreatePatient(t *testing.T) {
	beforeEach(t)

	gin.SetMode(gin.TestMode)
	tests := []struct {
		name           string
		body           string
		mockSetup      func(svc *mock.MockPatientService)
		wantStatusCode int
	}{
		{
			name: "성공",
			body: `{
				"patientId": "test",
				"name": "test",
				"gender": "M",
				"birthDate": "1990-01-01"
			}`,
			mockSetup: func(svc *mock.MockPatientService) {
				svc.EXPECT().
					CreatePatient(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "실패 - 필수 필드 누락",
			body: `{
				"patientId": "test"
			}`,
			mockSetup:      func(svc *mock.MockPatientService) {},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "실패 - 비즈니스 로직 에러 (500)",
			body: `{
				"patientId": "test",
				"name": "test",
				"gender": "M",
				"birthDate": "1990-01-01"
			}`,
			mockSetup: func(svc *mock.MockPatientService) {
				svc.EXPECT().
					CreatePatient(gomock.Any(), gomock.Any()).
					Return(pkgError.WrapWithCode(pkgError.EmptyBusinessError(), pkgError.Create))
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup(mockService)

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest(
				http.MethodPost,
				"/v1/patients",
				strings.NewReader(tt.body),
			)
			req.Header.Set("Content-Type", "application/json")
			ctx.Request = req

			controller.CreatePatient(ctx)

			require.Equal(t, tt.wantStatusCode, w.Code)
		})
	}
}

func Test_UpdatePatient(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name           string
		patientID      string
		body           string
		mockSetup      func(svc *mock.MockPatientService)
		wantStatusCode int
	}{
		{
			name:      "성공 - 환자 정보 수정",
			patientID: "P00001234",
			body: `{
				"name": "홍길동수정",
				"gender": "F",
				"birthDate": "1975-03-01",
				"version": 1
			}`,
			mockSetup: func(svc *mock.MockPatientService) {
				svc.EXPECT().
					UpdatePatient(gomock.Any(), "P00001234", gomock.Any()).
					Return(nil)
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:      "실패 - 필수 필드 누락 (version 없음)",
			patientID: "P00001234",
			body: `{
				"name": "홍길동수정",
				"gender": "F",
				"birthDate": "1975-03-01"
			}`,
			mockSetup:      func(svc *mock.MockPatientService) {},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:      "실패 - Version Conflict (Optimistic Lock)",
			patientID: "P00001234",
			body: `{
				"name": "홍길동수정",
				"gender": "F",
				"birthDate": "1975-03-01",
				"version": 2
			}`,
			mockSetup: func(svc *mock.MockPatientService) {
				svc.EXPECT().
					UpdatePatient(gomock.Any(), "P00001234", gomock.Any()).
					Return(pkgError.WrapWithCode(pkgError.EmptyBusinessError(), pkgError.Conflict, "version mismatch"))
			},
			wantStatusCode: http.StatusConflict,
		},
		{
			name:      "실패 - 잘못된 Gender 값",
			patientID: "P00001234",
			body: `{
				"name": "홍길동수정",
				"gender": "X",
				"birthDate": "1975-03-01",
				"version": 1
			}`,
			mockSetup:      func(svc *mock.MockPatientService) {},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:      "실패 - 잘못된 날짜 형식",
			patientID: "P00001234",
			body: `{
				"name": "홍길동수정",
				"gender": "F",
				"birthDate": "19750301",
				"version": 1
			}`,
			mockSetup:      func(svc *mock.MockPatientService) {},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:      "실패 - patient_id 파라미터 없음",
			patientID: "",
			body: `{
				"name": "홍길동수정",
				"gender": "F",
				"birthDate": "1975-03-01",
				"version": 1
			}`,
			mockSetup:      func(svc *mock.MockPatientService) {},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:      "실패 - 비즈니스 로직 에러 (500)",
			patientID: "P00001234",
			body: `{
				"name": "홍길동수정",
				"gender": "F",
				"birthDate": "1975-03-01",
				"version": 1
			}`,
			mockSetup: func(svc *mock.MockPatientService) {
				svc.EXPECT().
					UpdatePatient(gomock.Any(), "P00001234", gomock.Any()).
					Return(pkgError.WrapWithCode(pkgError.EmptyBusinessError(), pkgError.Update))
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeEach(t)
			tt.mockSetup(mockService)

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest(
				http.MethodPut,
				"/v1/patients/"+tt.patientID,
				strings.NewReader(tt.body),
			)
			req.Header.Set("Content-Type", "application/json")
			ctx.Request = req
			ctx.Params = gin.Params{
				{Key: "patient_id", Value: tt.patientID},
			}

			controller.UpdatePatient(ctx)

			require.Equal(t, tt.wantStatusCode, w.Code)
		})
	}
}

func Test_GetPatientVitals(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		patientID      string
		queryString    string
		mockSetup      func(svc *mock.MockPatientService)
		wantStatusCode int
	}{
		{
			name:        "성공 - vital_type 있을 때",
			patientID:   "P00001234",
			queryString: "from=2025-12-01T10:00:00Z&to=2025-12-01T12:00:00Z&vital_type=HR",
			mockSetup: func(svc *mock.MockPatientService) {
				svc.EXPECT().
					GetPatientVitals(gomock.Any(), "P00001234", gomock.Any()).
					Return(&patient.GetPatientVitalsResponse{
						PatientID: "P00001234",
						Items: []patient.VitalItemResponse{
							{VitalType: "HR", RecordedAt: time.Now(), Value: 110.0},
						},
					}, nil)
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:        "성공 - vital_type 없을 때 (모든 타입)",
			patientID:   "P00001234",
			queryString: "from=2025-12-01T10:00:00Z&to=2025-12-01T12:00:00Z",
			mockSetup: func(svc *mock.MockPatientService) {
				svc.EXPECT().
					GetPatientVitals(gomock.Any(), "P00001234", gomock.Any()).
					Return(&patient.GetPatientVitalsResponse{
						PatientID: "P00001234",
						Items: []patient.VitalItemResponse{
							{VitalType: "HR", RecordedAt: time.Now(), Value: 110.0},
							{VitalType: "RR", RecordedAt: time.Now(), Value: 20.0},
						},
					}, nil)
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "실패 - patient_id 파라미터 없음",
			patientID:      "",
			queryString:    "from=2025-12-01T10:00:00Z&to=2025-12-01T12:00:00Z&vital_type=HR",
			mockSetup:      func(svc *mock.MockPatientService) {},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "실패 - from 파라미터 없음",
			patientID:      "P00001234",
			queryString:    "to=2025-12-01T12:00:00Z&vital_type=HR",
			mockSetup:      func(svc *mock.MockPatientService) {},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "실패 - to 파라미터 없음",
			patientID:      "P00001234",
			queryString:    "from=2025-12-01T10:00:00Z&vital_type=HR",
			mockSetup:      func(svc *mock.MockPatientService) {},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "실패 - 잘못된 vital_type",
			patientID:      "P00001234",
			queryString:    "from=2025-12-01T10:00:00Z&to=2025-12-01T12:00:00Z&vital_type=INVALID",
			mockSetup:      func(svc *mock.MockPatientService) {},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:        "실패 - Service 에러 (잘못된 날짜 형식)",
			patientID:   "P00001234",
			queryString: "from=2025-12-01T10:00:00Z&to=2025-12-01T12:00:00Z&vital_type=HR",
			mockSetup: func(svc *mock.MockPatientService) {
				svc.EXPECT().
					GetPatientVitals(gomock.Any(), "P00001234", gomock.Any()).
					Return(nil, pkgError.WrapWithCode(pkgError.EmptyBusinessError(), pkgError.WrongParam, "invalid date format"))
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:        "실패 - Service 에러 (DB 조회 실패)",
			patientID:   "P00001234",
			queryString: "from=2025-12-01T10:00:00Z&to=2025-12-01T12:00:00Z&vital_type=HR",
			mockSetup: func(svc *mock.MockPatientService) {
				svc.EXPECT().
					GetPatientVitals(gomock.Any(), "P00001234", gomock.Any()).
					Return(nil, pkgError.WrapWithCode(pkgError.EmptyBusinessError(), pkgError.Get, "db error"))
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeEach(t)
			tt.mockSetup(mockService)

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest(
				http.MethodGet,
				"/v1/patients/"+tt.patientID+"/vitals?"+tt.queryString,
				nil,
			)
			ctx.Request = req
			if tt.patientID != "" {
				ctx.Params = gin.Params{
					{Key: "patient_id", Value: tt.patientID},
				}
			}

			controller.GetPatientVitals(ctx)

			require.Equal(t, tt.wantStatusCode, w.Code)
		})
	}
}
