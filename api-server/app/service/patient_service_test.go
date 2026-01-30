package service

import (
	"aitrics-vital-signs/api-server/domain/mock"
	"aitrics-vital-signs/api-server/domain/patient"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var (
	mockRepository *mock.MockPatientRepository
	svc            patient.PatientService
)

func beforeEach(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRepository = mock.NewMockPatientRepository(ctrl)
	svc = NewPatientService(mockRepository)
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
