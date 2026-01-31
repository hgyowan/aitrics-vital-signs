package service

import (
	"aitrics-vital-signs/api-server/domain/mock"
	"aitrics-vital-signs/api-server/domain/patient"
	"aitrics-vital-signs/api-server/domain/vital"
	pkgError "aitrics-vital-signs/library/error"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

var (
	mockRepository   *mock.MockPatientRepository
	mockVitalService *mock.MockVitalService
	svc              patient.PatientService
)

func beforeEach(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepository = mock.NewMockPatientRepository(ctrl)
	mockVitalService = mock.NewMockVitalService(ctrl)
	svc = NewPatientService(mockRepository, mockVitalService)
}

func Test_CreatePatient(t *testing.T) {
	beforeEach(t)

	tests := []struct {
		name    string
		req     patient.CreatePatientRequest
		wantErr bool
	}{
		{
			name: "성공",
			req: patient.CreatePatientRequest{
				PatientID: "test",
				Name:      "test",
				Gender:    "M",
				BirthDate: "1990-01-01",
			},
			wantErr: false,
		},
		{
			name: "실패 - 날짜 파라미터 포멧 에러",
			req: patient.CreatePatientRequest{
				PatientID: "test",
				Name:      "test",
				Gender:    "M",
				BirthDate: "19900101",
			},
			wantErr: true,
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepository.EXPECT().
				CreatePatient(ctx, gomock.Any()).
				DoAndReturn(func(_ context.Context, p *patient.Patient) error {
					require.NotEmpty(t, p.ID)
					require.Equal(t, tt.req.PatientID, p.PatientID)
					require.Equal(t, tt.req.Name, p.Name)
					require.Equal(t, tt.req.Gender, p.Gender)
					require.NotEmpty(t, p.BirthDate)
					require.NotEmpty(t, p.CreatedAt)
					require.NotEmpty(t, p.UpdatedAt)
					return nil
				}).AnyTimes()

			err := svc.CreatePatient(ctx, tt.req)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_UpdatePatient(t *testing.T) {
	tests := []struct {
		name        string
		patientID   string
		req         patient.UpdatePatientRequest
		setupMock   func()
		wantErr     bool
		expectedErr error
	}{
		{
			name:      "성공 - 환자 정보 수정",
			patientID: "P00001234",
			req: patient.UpdatePatientRequest{
				Name:      "홍길동수정",
				Gender:    "F",
				BirthDate: "1975-03-01",
				Version:   1,
			},
			setupMock: func() {
				now := time.Now().UTC()
				existingPatient := &patient.Patient{
					ID:        uuid.NewString(),
					PatientID: "P00001234",
					Name:      "홍길동",
					Gender:    "M",
					BirthDate: now,
					Version:   1,
					CreatedAt: now,
					UpdatedAt: &now,
				}
				mockRepository.EXPECT().
					FindPatientByID(gomock.Any(), "P00001234").
					Return(existingPatient, nil)
				mockRepository.EXPECT().
					UpdatePatient(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, p *patient.Patient) error {
						require.Equal(t, "홍길동수정", p.Name)
						require.Equal(t, "F", p.Gender)
						require.Equal(t, 2, p.Version)
						require.NotNil(t, p.UpdatedAt)
						return nil
					})
			},
			wantErr: false,
		},
		{
			name:      "실패 - Version Conflict (Optimistic Lock)",
			patientID: "P00001234",
			req: patient.UpdatePatientRequest{
				Name:      "홍길동수정",
				Gender:    "F",
				BirthDate: "1975-03-01",
				Version:   2,
			},
			setupMock: func() {
				now := time.Now().UTC()
				existingPatient := &patient.Patient{
					ID:        uuid.NewString(),
					PatientID: "P00001234",
					Name:      "홍길동",
					Gender:    "M",
					BirthDate: now,
					Version:   1,
					CreatedAt: now,
					UpdatedAt: &now,
				}
				mockRepository.EXPECT().
					FindPatientByID(gomock.Any(), "P00001234").
					Return(existingPatient, nil)
			},
			wantErr:     true,
			expectedErr: pkgError.WrapWithCode(pkgError.EmptyBusinessError(), pkgError.Conflict),
		},
		{
			name:      "실패 - 환자 없음",
			patientID: "P99999999",
			req: patient.UpdatePatientRequest{
				Name:      "홍길동수정",
				Gender:    "F",
				BirthDate: "1975-03-01",
				Version:   1,
			},
			setupMock: func() {
				mockRepository.EXPECT().
					FindPatientByID(gomock.Any(), "P99999999").
					Return(nil, pkgError.WrapWithCode(gorm.ErrRecordNotFound, pkgError.NotFound))
			},
			wantErr:     true,
			expectedErr: pkgError.WrapWithCode(gorm.ErrRecordNotFound, pkgError.NotFound),
		},
		{
			name:      "실패 - 날짜 파라미터 포멧 에러",
			patientID: "P00001234",
			req: patient.UpdatePatientRequest{
				Name:      "홍길동수정",
				Gender:    "F",
				BirthDate: "19750301",
				Version:   1,
			},
			setupMock: func() {},
			wantErr:   true,
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeEach(t)
			tt.setupMock()

			err := svc.UpdatePatient(ctx, tt.patientID, tt.req)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErr != nil {
					if pkgError.CompareBusinessError(tt.expectedErr, pkgError.Conflict) {
						require.True(t, pkgError.CompareBusinessError(err, pkgError.Conflict))
					} else if pkgError.CompareBusinessError(tt.expectedErr, pkgError.NotFound) {
						require.True(t, pkgError.CompareBusinessError(err, pkgError.NotFound))
					}
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_GetPatientVitals(t *testing.T) {
	tests := []struct {
		name      string
		patientID string
		req       patient.GetPatientVitalsQueryRequest
		setupMock func()
		wantErr   bool
	}{
		{
			name:      "성공 - vital_type 있을 때",
			patientID: "P00001234",
			req: patient.GetPatientVitalsQueryRequest{
				From:      "2025-12-01T10:00:00Z",
				To:        "2025-12-01T12:00:00Z",
				VitalType: "HR",
			},
			setupMock: func() {
				expectedResponse := &vital.GetVitalsResponse{
					PatientID: "P00001234",
					Items: []vital.VitalItemResponse{
						{
							VitalType:  "HR",
							RecordedAt: time.Date(2025, 12, 1, 10, 15, 0, 0, time.UTC),
							Value:      110.0,
						},
					},
				}
				mockVitalService.EXPECT().
					GetVitalsByPatientIDAndDateRange(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, req vital.GetVitalsRequest) (*vital.GetVitalsResponse, error) {
						require.Equal(t, "P00001234", req.PatientID)
						require.Equal(t, "HR", req.VitalType)
						return expectedResponse, nil
					})
			},
			wantErr: false,
		},
		{
			name:      "성공 - vital_type 없을 때 (모든 타입)",
			patientID: "P00001234",
			req: patient.GetPatientVitalsQueryRequest{
				From:      "2025-12-01T10:00:00Z",
				To:        "2025-12-01T12:00:00Z",
				VitalType: "",
			},
			setupMock: func() {
				expectedResponse := &vital.GetVitalsResponse{
					PatientID: "P00001234",
					Items: []vital.VitalItemResponse{
						{
							VitalType:  "HR",
							RecordedAt: time.Date(2025, 12, 1, 10, 15, 0, 0, time.UTC),
							Value:      110.0,
						},
						{
							VitalType:  "RR",
							RecordedAt: time.Date(2025, 12, 1, 10, 15, 0, 0, time.UTC),
							Value:      20.0,
						},
					},
				}
				mockVitalService.EXPECT().
					GetVitalsByPatientIDAndDateRange(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, req vital.GetVitalsRequest) (*vital.GetVitalsResponse, error) {
						require.Equal(t, "P00001234", req.PatientID)
						require.Equal(t, "", req.VitalType)
						return expectedResponse, nil
					})
			},
			wantErr: false,
		},
		{
			name:      "실패 - 잘못된 from 날짜 형식",
			patientID: "P00001234",
			req: patient.GetPatientVitalsQueryRequest{
				From:      "2025-12-01 10:00:00",
				To:        "2025-12-01T12:00:00Z",
				VitalType: "HR",
			},
			setupMock: func() {},
			wantErr:   true,
		},
		{
			name:      "실패 - 잘못된 to 날짜 형식",
			patientID: "P00001234",
			req: patient.GetPatientVitalsQueryRequest{
				From:      "2025-12-01T10:00:00Z",
				To:        "2025-12-01 12:00:00",
				VitalType: "HR",
			},
			setupMock: func() {},
			wantErr:   true,
		},
		{
			name:      "실패 - Vital Service 에러",
			patientID: "P00001234",
			req: patient.GetPatientVitalsQueryRequest{
				From:      "2025-12-01T10:00:00Z",
				To:        "2025-12-01T12:00:00Z",
				VitalType: "HR",
			},
			setupMock: func() {
				mockVitalService.EXPECT().
					GetVitalsByPatientIDAndDateRange(gomock.Any(), gomock.Any()).
					Return(nil, pkgError.WrapWithCode(pkgError.EmptyBusinessError(), pkgError.Get, "db error"))
			},
			wantErr: true,
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeEach(t)
			tt.setupMock()

			result, err := svc.GetPatientVitals(ctx, tt.patientID, tt.req)

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
