package main

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/Alexchent/ad-track/config"
	"github.com/Alexchent/ad-track/middleware"
	"github.com/fvbock/endless"
	"github.com/zeromicro/go-zero/core/conf"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"gopkg.in/natefinch/lumberjack.v2"
)

var configFile = flag.String("f", "conf.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	setupLogger(c.Log)

	router := gin.New()
	router.Use(requestid.New())
	router.Use(middleware.RequestLogger())
	//router.Use(middleware.PrometheusMetrics())
	router.Use(gin.Recovery())

	// 初始化并应用限流中间件
	//if c.RateLimit.Enabled {
	//	middleware.InitRateLimiter(c.RateLimit.Rate, c.RateLimit.Capacity)
	//	router.Use(middleware.RateLimit())
	//}
	register(router)

	// 使用 endless 实现平滑重启
	// 支持 SIGHUP 信号进行平滑重启，SIGTERM/SIGINT 信号进行优雅关闭
	server := endless.NewServer(c.Port, router)

	server.BeforeBegin = func(add string) {
		fmt.Println(fmt.Sprintf("服务启动, pid=%d, adder=%s", os.Getpid(), add))
	}

	if err := server.ListenAndServe(); err != nil {
		// 优雅关闭时，listener 会被主动关闭，此时 accept 会返回 "use of closed network connection" 错误
		// 这是正常行为，不应该记录为 fatal
		if strings.Contains(err.Error(), "use of closed network connection") {
			fmt.Println("服务已优雅关闭")
		} else {
			fmt.Println("服务错误")
		}
	}
}

func setupLogger(conf config.Logger) {
	// 配置日志轮转
	logRotate := &lumberjack.Logger{
		Filename:   conf.Filename,
		MaxSize:    conf.MaxSize, // MB
		MaxBackups: 3,
		MaxAge:     conf.MaxAge, // days
		Compress:   conf.Compress,
	}

	logLevel := slog.LevelInfo
	switch strings.ToLower(conf.Level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	case "fatal":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	// 自定义日志格式
	opts := &slog.HandlerOptions{
		Level: logLevel,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// 自定义时间格式
			if a.Key == slog.TimeKey {
				if t, ok := a.Value.Any().(time.Time); ok {
					a.Value = slog.StringValue(t.Format(time.RFC3339))
				}
			}
			return a
		},
	}

	// 同时输出到文件和控制台
	multiWriter := io.MultiWriter(logRotate, os.Stdout)

	// 创建JSON格式的logger
	var handler slog.Handler
	switch strings.ToLower(conf.Encoding) {
	case "console":
		handler = slog.NewTextHandler(multiWriter, opts)
	case "json":
		handler = slog.NewJSONHandler(multiWriter, opts)
	default:
		handler = slog.NewJSONHandler(multiWriter, opts)
	}
	logger := slog.New(handler)
	slog.SetDefault(logger)
}
