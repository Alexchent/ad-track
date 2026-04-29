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

	router.GET("/vivo/auth", handler.GetAuthorizationCode(svcCtx))
	// 监测链接
	router.Any("/click", handler.ProcessClick(svcCtx))
	router.POST("/vivo/click", handler.ProcessVIVOClick(svcCtx))
	// 归因
	router.GET("/report", handler.AttributeReport(svcCtx))
}

func customizedRegister(router *gin.Engine) {
	router.GET("/health", func(c *gin.Context) {
		slog.InfoContext(c.Request.Context(), "health check")
		c.String(http.StatusOK, "ok")
	})

	// Prometheus 指标接口
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
}
