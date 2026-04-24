package middleware

import (
	"ai-server/app"
	"time"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RequestLogger 请求日志中间件，增加链路追踪
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		method := c.Request.Method
		clientIP := c.ClientIP()

		// 获取请求ID作为链路追踪ID
		traceID := requestid.Get(c)

		// 将traceID设置到gin上下文中，供后续handler使用
		c.Set(string(app.TraceIDKey), traceID)

		// 将traceID设置到request context中，供repository等层使用
		ctx := app.SetTraceID(c.Request.Context(), traceID)
		c.Request = c.Request.WithContext(ctx)

		// 请求开始日志
		app.GetLogger().With(zap.String("trace_id", traceID))
		if query != "" {
			app.LogInfo(c.Request.Context(), "请求开始",
				zap.String("method", method),
				zap.String("path", path),
				zap.String("query", query),
				zap.String("client_ip", clientIP))
		} else {
			app.LogInfo(c.Request.Context(), "请求开始",
				zap.String("method", method),
				zap.String("path", path),
				zap.String("client_ip", clientIP))
		}

		// 处理请求
		c.Next()

		// 请求结束日志
		latency := time.Since(start)
		status := c.Writer.Status()

		app.LogInfo(c.Request.Context(), "请求结束",
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("latency", latency))
	}
}
