package svc

import (
	"github.com/Alexchent/ad-track/config"
	"github.com/Alexchent/ad-track/pkg/vivo"
)

type ServiceContext struct {
	Config    config.Config
	AdService *vivo.AdService
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
		AdService: vivo.NewAdService(&vivo.Config{
			Host:         c.VIVO.Host,
			ClientId:     c.VIVO.ClientId,
			ClientSecret: c.VIVO.ClientSecret,
		}),
	}
}
