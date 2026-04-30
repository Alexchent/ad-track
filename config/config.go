package config

type Config struct {
	//rest.RestConf
	App
	Redis RedisConfig
	MySQL MySQLConfig
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

type MySQLConfig struct {
	DSN string `json:"dsn"`
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
	ClientId     string            `json:"client_id"`
	ClientSecret string            `json:"client_secret"`
	APP          map[string]string `json:"APP"`
}

//type VAPP struct {
//	PkgName string `json:"pkgName"`
//	SrcId   string `json:"srcId"`
//}
