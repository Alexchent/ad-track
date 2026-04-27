package main

import (
	"log/slog"
	"net/http"

	"github.com/Alexchent/ad-track/handler"
	"github.com/Alexchent/ad-track/svc"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func register(router *gin.Engine, svcCtx *svc.ServiceContext) {
	customizedRegister(router)

	vivo := router.Group("/vivo")
	vivo.GET("auth", handler.GetAuthorizationCode(svcCtx))
	// 监测链接
	vivo.POST("click", handler.ProcessVIVOClick)
	// 归因
	router.GET("/report", func(context *gin.Context) {
		// 通过设备信息，如：oaid 等匹配监测到的点击数据，判断归因的媒体，最后将激活、留存等行为上报给对应的媒体
	})
}

func customizedRegister(router *gin.Engine) {
	router.GET("/health", func(c *gin.Context) {
		slog.InfoContext(c.Request.Context(), "health check")
		c.String(http.StatusOK, "ok")
	})

	// Prometheus 指标接口
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
}
