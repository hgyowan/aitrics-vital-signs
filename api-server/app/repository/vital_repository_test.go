package repository

import (
	"aitrics-vital-signs/api-server/domain/mock"
	"aitrics-vital-signs/api-server/domain/vital"
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var vitalRepo vital.VitalRepository
var vitalSQLMock sqlmock.Sqlmock

func beforeEachVital(t *testing.T) {
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
	vitalRepo = NewVitalRepository(mockExternalDBClient)
	vitalSQLMock = mockSQL
}

func Test_FindVitalByPatientIDAndRecordedAtAndVitalType(t *testing.T) {
	tests := []struct {
		name       string
		patientID  string
		recordedAt time.Time
		vitalType  string
		setupMock  func()
		wantErr    bool
	}{
		{
			name:       "성공 - Vital 조회",
			patientID:  "P00001234",
			recordedAt: time.Date(2025, 12, 1, 10, 15, 0, 0, time.UTC),
			vitalType:  "HR",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"patient_id", "recorded_at", "vital_type", "value", "version", "created_at"}).
					AddRow("P00001234", time.Date(2025, 12, 1, 10, 15, 0, 0, time.UTC), "HR", 110.0, 1, time.Now().UTC())
				vitalSQLMock.ExpectQuery("SELECT .*FROM .*vitals.* WHERE patient_id = .* AND recorded_at = .* AND vital_type = .*").
					WithArgs("P00001234", time.Date(2025, 12, 1, 10, 15, 0, 0, time.UTC), "HR", 1).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name:       "실패 - Vital 없음",
			patientID:  "P99999999",
			recordedAt: time.Date(2025, 12, 1, 10, 15, 0, 0, time.UTC),
			vitalType:  "HR",
			setupMock: func() {
				vitalSQLMock.ExpectQuery("SELECT .*FROM .*vitals.* WHERE patient_id = .* AND recorded_at = .* AND vital_type = .*").
					WithArgs("P99999999", time.Date(2025, 12, 1, 10, 15, 0, 0, time.UTC), "HR", 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeEachVital(t)
			tt.setupMock()

			result, err := vitalRepo.FindVitalByPatientIDAndRecordedAtAndVitalType(context.Background(), tt.patientID, tt.recordedAt, tt.vitalType)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Equal(t, tt.patientID, result.PatientID)
				require.Equal(t, tt.vitalType, result.VitalType)
			}
		})
	}
}

func Test_CreateVital(t *testing.T) {
	beforeEachVital(t)

	now := time.Now().UTC()
	vitalSQLMock.ExpectBegin()
	vitalSQLMock.ExpectExec("INSERT INTO .*vitals.*").
		WillReturnResult(sqlmock.NewResult(1, 1))
	vitalSQLMock.ExpectCommit()

	err := vitalRepo.CreateVital(context.Background(), &vital.Vital{
		PatientID:  "P00001234",
		RecordedAt: time.Date(2025, 12, 1, 10, 15, 0, 0, time.UTC),
		VitalType:  "HR",
		Value:      110.0,
		Version:    1,
		CreatedAt:  now,
		UpdatedAt:  &now,
	})
	require.NoError(t, err)
}

func Test_UpdateVital(t *testing.T) {
	beforeEachVital(t)

	now := time.Now().UTC()
	updateModel := &vital.Vital{
		PatientID:  "P00001234",
		RecordedAt: time.Date(2025, 12, 1, 10, 15, 0, 0, time.UTC),
		VitalType:  "HR",
		Value:      120.0,
		Version:    2,
		CreatedAt:  now,
		UpdatedAt:  &now,
	}

	vitalSQLMock.ExpectBegin()
	vitalSQLMock.ExpectExec("UPDATE .*vitals.*").
		WillReturnResult(sqlmock.NewResult(1, 1))
	vitalSQLMock.ExpectCommit()

	err := vitalRepo.UpdateVital(context.Background(), updateModel)
	require.NoError(t, err)
}

func Test_FindVitalsByPatientIDAndDateRange(t *testing.T) {
	from := time.Date(2025, 12, 1, 10, 0, 0, 0, time.UTC)
	to := time.Date(2025, 12, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		patientID string
		from      time.Time
		to        time.Time
		vitalType string
		setupMock func()
		wantCount int
		wantErr   bool
	}{
		{
			name:      "성공 - vital_type 있을 때 해당 타입만 조회",
			patientID: "P00001234",
			from:      from,
			to:        to,
			vitalType: "HR",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"patient_id", "recorded_at", "vital_type", "value", "version", "created_at"}).
					AddRow("P00001234", time.Date(2025, 12, 1, 10, 15, 0, 0, time.UTC), "HR", 110.0, 1, time.Now().UTC()).
					AddRow("P00001234", time.Date(2025, 12, 1, 11, 15, 0, 0, time.UTC), "HR", 115.0, 1, time.Now().UTC())
				vitalSQLMock.ExpectQuery("SELECT .* FROM .*vitals.* WHERE .* ORDER BY recorded_at").
					WithArgs("P00001234", from, to, "HR").
					WillReturnRows(rows)
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "성공 - vital_type 없을 때 모든 타입 조회",
			patientID: "P00001234",
			from:      from,
			to:        to,
			vitalType: "",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"patient_id", "recorded_at", "vital_type", "value", "version", "created_at"}).
					AddRow("P00001234", time.Date(2025, 12, 1, 10, 15, 0, 0, time.UTC), "HR", 110.0, 1, time.Now().UTC()).
					AddRow("P00001234", time.Date(2025, 12, 1, 10, 15, 0, 0, time.UTC), "RR", 20.0, 1, time.Now().UTC()).
					AddRow("P00001234", time.Date(2025, 12, 1, 11, 15, 0, 0, time.UTC), "HR", 115.0, 1, time.Now().UTC())
				vitalSQLMock.ExpectQuery("SELECT .* FROM .*vitals.* WHERE .* ORDER BY recorded_at").
					WithArgs("P00001234", from, to).
					WillReturnRows(rows)
			},
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:      "성공 - 조회 결과 없음",
			patientID: "P99999999",
			from:      from,
			to:        to,
			vitalType: "HR",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"patient_id", "recorded_at", "vital_type", "value", "version", "created_at"})
				vitalSQLMock.ExpectQuery("SELECT .* FROM .*vitals.* WHERE .* ORDER BY recorded_at").
					WithArgs("P99999999", from, to, "HR").
					WillReturnRows(rows)
			},
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeEachVital(t)
			tt.setupMock()

			results, err := vitalRepo.FindVitalsByPatientIDAndDateRange(context.Background(), tt.patientID, tt.from, tt.to, tt.vitalType)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, results)
			} else {
				require.NoError(t, err)
				require.NotNil(t, results)
				require.Equal(t, tt.wantCount, len(results))
			}
		})
	}
}
