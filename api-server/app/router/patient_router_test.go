package router

import (
	"aitrics-vital-signs/api-server/domain/mock"
	"aitrics-vital-signs/library/envs"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_ValidToken(t *testing.T) {
	t.Setenv("TOKEN", "test-token-123")
	envs.Token = os.Getenv("TOKEN")
	gin.SetMode(gin.TestMode)

	engine := gin.New()

	ctrl := gomock.NewController(t)
	patientController := mock.NewMockPatientController(ctrl)
	NewPatientRouter(engine, patientController)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/patients",
		nil,
	)
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}
