package controller

import (
	"aitrics-vital-signs/api-server/domain/mock"
	"aitrics-vital-signs/api-server/domain/patient"
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
