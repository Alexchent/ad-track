package svc

import (
	"github.com/Alexchent/ad-track/config"
	"github.com/Alexchent/ad-track/pkg/vivo"
	"github.com/go-redis/redis/v8"
)

type ServiceContext struct {
	Config    config.Config
	AdService *vivo.AdService
}

func NewServiceContext(c config.Config) *ServiceContext {
	cacheClient := redis.NewClient(&redis.Options{Addr: c.Redis.Addr, Password: c.Redis.Password, DB: c.Redis.Db})
	return &ServiceContext{
		Config: c,
		AdService: vivo.NewAdService(&vivo.Config{
			Host:         c.VIVO.Host,
			ClientId:     c.VIVO.ClientId,
			ClientSecret: c.VIVO.ClientSecret,
		}, cacheClient),
	}
}
