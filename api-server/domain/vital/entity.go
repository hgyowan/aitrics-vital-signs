package vital

import (
	"time"

	"gorm.io/gorm"
)

type Vital struct {
	PatientID  string         `gorm:"column:patient_id;type:varchar(20);not null;primaryKey;comment:외부 환자 ID"`
	RecordedAt time.Time      `gorm:"column:recorded_at;type:datetime(3);not null;primaryKey;comment:레코드 기록일"`
	VitalType  string         `gorm:"column:vital_type;type:enum('HR', 'RR', 'SBP', 'DBP', 'SpO2', 'BT');not null;comment:바이탈 유형"`
	Value      float64        `gorm:"column:birth_date;type:double;not null;comment:바이탈 값"`
	Version    int            `gorm:"column:version;not null;default:1;comment:버전"`
	CreatedAt  time.Time      `gorm:"column:created_at;type:datetime(3);not null;comment:데이터 생성일"`
	UpdatedAt  *time.Time     `gorm:"column:updated_at;type:datetime(3);comment:데이터 수정일"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;type:datetime(3);comment:데이터 삭제일"`
}

func (v *Vital) TableName() string {
	return "vitals"
}
