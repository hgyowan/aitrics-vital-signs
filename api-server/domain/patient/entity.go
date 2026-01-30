package patient

import (
	"time"

	"gorm.io/gorm"
)

type Patient struct {
	ID        string         `gorm:"column:id;type:char(36);primaryKey;comment:PK"`
	PatientID string         `gorm:"column:patient_id;type:varchar(20);not null;uniqueIndex;comment:외부 환자 ID"`
	Name      string         `gorm:"column:name;type:varchar(50);not null;comment:환자 이름"`
	Gender    string         `gorm:"column:gender;type:enum('M','F');not null;comment:성별"`
	BirthDate time.Time      `gorm:"column:birth_date;type:date;not null;comment:생년월일"`
	Version   int            `gorm:"column:version;not null;default:1;comment:버전"`
	CreatedAt time.Time      `gorm:"column:created_at;type:datetime(3);not null;comment:데이터 생성일"`
	UpdatedAt *time.Time     `gorm:"column:updated_at;type:datetime(3);comment:데이터 수정일"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;type:datetime(3);comment:데이터 삭제일"`
}

func (p *Patient) TableName() string {
	return "patients"
}
