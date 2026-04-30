package config

type Config struct {
	//rest.RestConf
	App
	Redis RedisConfig
	Log   Logger
	VIVO  VIVOConfig
}

type App struct {
	Port        string
	Env         string `json:",optional"`
	CachePrefix string `json:",optional"`
}

type RedisConfig struct {
	Addr     string `json:"addr"`
	Password string `json:"password"`
	Db       int    `json:"db"`
}

type ESConfig struct {
	Addresses []string `json:"addresses"`
	Index     string   `json:"index"`
}

type Logger struct {
	Filename string
	Encoding string `json:",optional"`
	Level    string `json:",optional"`
	MaxSize  int
	MaxAge   int
	Compress bool
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Rate     int64 `json:"rate"`     // 每秒允许的请求数
	Capacity int64 `json:"capacity"` // 桶的容量（允许的突发流量）
	Enabled  bool  `json:"enabled"`  // 是否启用限流
}

type VIVOConfig struct {
	Host         string            `json:"host"`
	ClientId     string            `json:"client_id"`     // 从 vivo 开放平台https://open-ad.vivo.com.cn/获取
	ClientSecret string            `json:"client_secret"` // 从 vivo 开放平台https://open-ad.vivo.com.cn/获取
	APP          map[string]string `json:"APP"`           // 授权的广告主下，事件源，从vivo营销平台https://ad.vivo.com.cn/marketing/property/event-manage 获取
}
