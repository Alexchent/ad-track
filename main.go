package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/Alexchent/ad-track/config"
	"github.com/Alexchent/ad-track/middleware"
	"github.com/Alexchent/ad-track/pkg/logger"
	"github.com/Alexchent/ad-track/svc"
	"github.com/fvbock/endless"
	"github.com/zeromicro/go-zero/core/conf"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

var configFile = flag.String("f", "conf.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	//setupLogger(c.Log)
	logger.SetupLogger(&logger.LogConfig{
		Filename: c.Log.Filename,
		Encoding: c.Log.Encoding,
		Level:    c.Log.Level,
		MaxSize:  c.Log.MaxSize,
		MaxAge:   c.Log.MaxAge,
		Compress: c.Log.Compress,
	})

	svcCtx := svc.NewServiceContext(c)

	router := gin.New()
	router.Use(requestid.New())
	router.Use(middleware.RequestLogger())
	router.Use(gin.Recovery())
	register(router, svcCtx)

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
