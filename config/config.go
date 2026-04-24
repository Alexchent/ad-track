package config

type Config struct {
	//rest.RestConf
	App
	Redis RedisConfig
	MySQL MySQLConfig
	Log   Logger
	//RateLimit RateLimitConfig // 限流配置
	VIVO VIVOConfig
}

type App struct {
	Port string
	Env  string `json:",optional"`
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
	// Encoding sets the logger's encoding. Valid values are "json" and
	// "console", as well as any third-party encodings registered via
	// RegisterEncoder.
	Encoding string `json:",optional"`
	Level    string `json:",optional"`
	MaxSize  int
	MaxAge   int
	Compress bool
	// RotateStrategy 日志轮转策略: "rotatelogs" 或 "lumberjack"
	// - rotatelogs: 使用 github.com/lestrrat-go/file-rotatelogs，按天轮转
	// - lumberjack: 使用 gopkg.in/natefinch/lumberjack.v2，按大小轮转
	RotateStrategy string `json:",optional"`
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Rate     int64 `json:"rate"`     // 每秒允许的请求数
	Capacity int64 `json:"capacity"` // 桶的容量（允许的突发流量）
	Enabled  bool  `json:"enabled"`  // 是否启用限流
}

type VIVOConfig struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}
