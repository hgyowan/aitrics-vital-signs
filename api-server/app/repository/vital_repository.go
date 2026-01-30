package repository

import (
	"aitrics-vital-signs/api-server/domain"
	"aitrics-vital-signs/api-server/domain/vital"
)

type vitalRepository struct {
	externalGormClient domain.ExternalDBClient
}

func NewVitalRepository(externalGormClient domain.ExternalDBClient) vital.VitalRepository {
	return &vitalRepository{externalGormClient: externalGormClient}
}
