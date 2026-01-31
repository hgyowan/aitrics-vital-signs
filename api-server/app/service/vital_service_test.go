package service

import (
	"aitrics-vital-signs/api-server/domain/mock"
	"aitrics-vital-signs/api-server/domain/vital"
	pkgError "aitrics-vital-signs/library/error"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var (
	mockVitalRepository *mock.MockVitalRepository
	vitalSvc            vital.VitalService
)

func beforeEachVital(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockVitalRepository = mock.NewMockVitalRepository(ctrl)
	vitalSvc = NewVitalService(mockVitalRepository)
}

func Test_UpsertVital_Insert(t *testing.T) {
	beforeEachVital(t)

	recordedAt := time.Date(2025, 12, 1, 10, 15, 0, 0, time.UTC)

	tests := []struct {
		name        string
		req         vital.UpsertVitalRequest
		setupMock   func()
		wantErr     bool
		expectedErr error
	}{
		{
			name: "성공 - INSERT (새 데이터)",
			req: vital.UpsertVitalRequest{
				PatientID:  "P00001234",
				RecordedAt: recordedAt,
				VitalType:  "HR",
				Value:      110.0,
				Version:    1,
			},
			setupMock: func() {
				// FindVital → Record Not Found
				mockVitalRepository.EXPECT().
					FindVitalByPatientIDAndRecordedAtAndVitalType(
						gomock.Any(), vital.FindVitalByPatientIDAndRecordedAtAndVitalTypeParam{
							PatientID:  "P00001234",
							RecordedAt: recordedAt,
							VitalType:  "HR",
						}).
					Return(nil, pkgError.WrapWithCode(pkgError.EmptyBusinessError(), pkgError.NotFound))

				// CreateVital 호출
				mockVitalRepository.EXPECT().
					CreateVital(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, v *vital.Vital) error {
						require.Equal(t, "P00001234", v.PatientID)
						require.Equal(t, recordedAt, v.RecordedAt)
						require.Equal(t, "HR", v.VitalType)
						require.Equal(t, 110.0, v.Value)
						require.Equal(t, 1, v.Version)
						require.NotNil(t, v.CreatedAt)
						require.NotNil(t, v.UpdatedAt)
						return nil
					})
			},
			wantErr: false,
		},
		{
			name: "실패 - INSERT 시 version이 1이 아님",
			req: vital.UpsertVitalRequest{
				PatientID:  "P00001234",
				RecordedAt: recordedAt,
				VitalType:  "HR",
				Value:      110.0,
				Version:    5,
			},
			setupMock: func() {
				mockVitalRepository.EXPECT().
					FindVitalByPatientIDAndRecordedAtAndVitalType(
						gomock.Any(), vital.FindVitalByPatientIDAndRecordedAtAndVitalTypeParam{
							PatientID:  "P00001234",
							RecordedAt: recordedAt,
							VitalType:  "HR",
						}).
					Return(nil, pkgError.WrapWithCode(pkgError.EmptyBusinessError(), pkgError.NotFound))
			},
			wantErr:     true,
			expectedErr: pkgError.WrapWithCode(pkgError.EmptyBusinessError(), pkgError.WrongParam),
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeEachVital(t)
			tt.setupMock()

			err := vitalSvc.UpsertVital(ctx, tt.req)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErr != nil {
					expectedBE, _ := pkgError.CastBusinessError(tt.expectedErr)
					actualBE, _ := pkgError.CastBusinessError(err)
					require.Equal(t, expectedBE.Status.Code, actualBE.Status.Code)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_UpsertVital_Update(t *testing.T) {
	beforeEachVital(t)

	recordedAt := time.Date(2025, 12, 1, 10, 15, 0, 0, time.UTC)
	now := time.Now().UTC()

	tests := []struct {
		name        string
		req         vital.UpsertVitalRequest
		setupMock   func()
		wantErr     bool
		expectedErr error
	}{
		{
			name: "성공 - UPDATE (기존 데이터)",
			req: vital.UpsertVitalRequest{
				PatientID:  "P00001234",
				RecordedAt: recordedAt,
				VitalType:  "HR",
				Value:      115.0,
				Version:    1,
			},
			setupMock: func() {
				existingVital := &vital.Vital{
					PatientID:  "P00001234",
					RecordedAt: recordedAt,
					VitalType:  "HR",
					Value:      110.0,
					Version:    1,
					CreatedAt:  now,
					UpdatedAt:  &now,
				}

				mockVitalRepository.EXPECT().
					FindVitalByPatientIDAndRecordedAtAndVitalType(
						gomock.Any(), vital.FindVitalByPatientIDAndRecordedAtAndVitalTypeParam{
							PatientID:  "P00001234",
							RecordedAt: recordedAt,
							VitalType:  "HR",
						}).
					Return(existingVital, nil)

				mockVitalRepository.EXPECT().
					UpdateVital(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, v *vital.Vital) error {
						require.Equal(t, 115.0, v.Value)
						require.Equal(t, 2, v.Version)
						require.NotNil(t, v.UpdatedAt)
						return nil
					})
			},
			wantErr: false,
		},
		{
			name: "실패 - Version Conflict (Optimistic Lock)",
			req: vital.UpsertVitalRequest{
				PatientID:  "P00001234",
				RecordedAt: recordedAt,
				VitalType:  "HR",
				Value:      120.0,
				Version:    1,
			},
			setupMock: func() {
				existingVital := &vital.Vital{
					PatientID:  "P00001234",
					RecordedAt: recordedAt,
					VitalType:  "HR",
					Value:      115.0,
					Version:    2, // DB의 version은 이미 2
					CreatedAt:  now,
					UpdatedAt:  &now,
				}

				mockVitalRepository.EXPECT().
					FindVitalByPatientIDAndRecordedAtAndVitalType(
						gomock.Any(), vital.FindVitalByPatientIDAndRecordedAtAndVitalTypeParam{
							PatientID:  "P00001234",
							RecordedAt: recordedAt,
							VitalType:  "HR",
						}).
					Return(existingVital, nil)
			},
			wantErr:     true,
			expectedErr: pkgError.WrapWithCode(pkgError.EmptyBusinessError(), pkgError.Conflict),
		},
		{
			name: "실패 - Repository에서 Version Conflict",
			req: vital.UpsertVitalRequest{
				PatientID:  "P00001234",
				RecordedAt: recordedAt,
				VitalType:  "HR",
				Value:      120.0,
				Version:    1,
			},
			setupMock: func() {
				existingVital := &vital.Vital{
					PatientID:  "P00001234",
					RecordedAt: recordedAt,
					VitalType:  "HR",
					Value:      110.0,
					Version:    1,
					CreatedAt:  now,
					UpdatedAt:  &now,
				}

				mockVitalRepository.EXPECT().
					FindVitalByPatientIDAndRecordedAtAndVitalType(
						gomock.Any(), vital.FindVitalByPatientIDAndRecordedAtAndVitalTypeParam{
							PatientID:  "P00001234",
							RecordedAt: recordedAt,
							VitalType:  "HR",
						}).
					Return(existingVital, nil)

				// Repository에서 Conflict 반환 (DB level 동시성 제어)
				mockVitalRepository.EXPECT().
					UpdateVital(gomock.Any(), gomock.Any()).
					Return(pkgError.WrapWithCode(pkgError.EmptyBusinessError(), pkgError.Conflict, "version conflict in db update"))
			},
			wantErr:     true,
			expectedErr: pkgError.WrapWithCode(pkgError.EmptyBusinessError(), pkgError.Conflict),
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeEachVital(t)
			tt.setupMock()

			err := vitalSvc.UpsertVital(ctx, tt.req)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErr != nil {
					if pkgError.CompareBusinessError(tt.expectedErr, pkgError.Conflict) {
						require.True(t, pkgError.CompareBusinessError(err, pkgError.Conflict))
					}
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
