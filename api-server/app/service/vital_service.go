package service

import (
	"aitrics-vital-signs/api-server/domain/vital"
)

type vitalService struct {
	repo vital.VitalRepository
}

func NewVitalService(repo vital.VitalRepository) vital.VitalService {
	return &vitalService{repo}
}
