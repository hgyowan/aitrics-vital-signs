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

func Test_FindPatientByID(t *testing.T) {
	tests := []struct {
		name      string
		patientID string
		setupMock func()
		wantErr   bool
	}{
		{
			name:      "성공 - 환자 조회",
			patientID: "P00001234",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "patient_id", "name", "gender", "birth_date", "version", "created_at"}).
					AddRow(uuid.NewString(), "P00001234", "홍길동", "M", time.Now().UTC(), 1, time.Now().UTC())
				sqlMock.ExpectQuery("SELECT .*FROM .*patients.* WHERE patient_id = .*").
					WithArgs("P00001234", 1).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name:      "실패 - 환자 없음",
			patientID: "P99999999",
			setupMock: func() {
				sqlMock.ExpectQuery("SELECT .*FROM .*patients.* WHERE patient_id = .*").
					WithArgs("P99999999", 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeEach(t)
			tt.setupMock()

			result, err := repo.FindPatientByID(context.Background(), tt.patientID)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Equal(t, tt.patientID, result.PatientID)
			}
		})
	}
}

func Test_UpdatePatient(t *testing.T) {
	beforeEach(t)

	now := time.Now().UTC()
	updateModel := &patient.Patient{
		ID:        uuid.NewString(),
		PatientID: "P00001234",
		Name:      "홍길동수정",
		Gender:    "F",
		BirthDate: now,
		Version:   2,
		CreatedAt: now,
		UpdatedAt: &now,
	}

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec("UPDATE .*patients.*").
		WillReturnResult(sqlmock.NewResult(1, 1))
	sqlMock.ExpectCommit()

	err := repo.UpdatePatient(context.Background(), updateModel)
	require.NoError(t, err)
}
