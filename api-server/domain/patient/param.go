package patient

type CreatePatientRequest struct {
	PatientID string `json:"patientId" binding:"required"`
	Name      string `json:"name" binding:"required"`
	Gender    string `json:"gender" binding:"required,oneof=M F"`
	BirthDate string `json:"birthDate" binding:"required,datetime=2006-01-02"`
}
