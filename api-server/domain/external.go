//go:generate mockgen -source=external.go -destination=./mock/mock_external.go -package=mock
package domain

import (
	"gorm.io/gorm"
)

type ExternalDBClient interface {
	MySQL() *gorm.DB
}
