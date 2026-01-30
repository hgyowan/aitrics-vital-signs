package repository

import (
	"aitrics-vital-signs/api-server/domain/mock"
	"aitrics-vital-signs/api-server/domain/patient"
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var repo patient.PatientRepository
var sqlMock sqlmock.Sqlmock

func beforeEach(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockExternalDBClient := mock.NewMockExternalDBClient(ctrl)

	sqlDB, mockSQL, err := sqlmock.New()
	require.NoError(t, err)

	dial := mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	})
	db, err := gorm.Open(dial, &gorm.Config{})
	require.NoError(t, err)

	mockExternalDBClient.EXPECT().MySQL().Return(db).AnyTimes()
	repo = NewPatientRepository(mockExternalDBClient)
	sqlMock = mockSQL
}

func Test_CreatePatient(t *testing.T) {
	beforeEach(t)

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec("INSERT INTO .*patients.*").
		WillReturnResult(sqlmock.NewResult(1, 1))
	sqlMock.ExpectCommit()
	err := repo.CreatePatient(context.Background(), &patient.Patient{
		ID:        uuid.NewString(),
		PatientID: "P00001234",
		Name:      "test",
		Gender:    "M",
		BirthDate: time.Now().UTC(),
		Version:   0,
		CreatedAt: time.Now().UTC(),
	})
	require.NoError(t, err)
}
