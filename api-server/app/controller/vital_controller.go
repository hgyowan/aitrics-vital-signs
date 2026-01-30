package controller

import (
	"aitrics-vital-signs/api-server/domain/vital"
)

type vitalController struct {
	service vital.VitalService
}

func NewVitalController(service vital.VitalService) vital.VitalRepository {
	p := &vitalController{
		service: service,
	}

	return p
}
