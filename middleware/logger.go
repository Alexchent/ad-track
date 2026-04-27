package middleware

import (
	"context"
	"log/slog"
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
		TraceIDKey := "traceID"
		c.Set(TraceIDKey, traceID)

		// 将traceID设置到request context中，供repository等层使用
		ctx := context.WithValue(c.Request.Context(), TraceIDKey, traceID)
		c.Request = c.Request.WithContext(ctx)

		// 请求开始日志
		traceLog := slog.With(zap.String("trace_id", traceID))
		slog.SetDefault(traceLog)

		// 处理请求
		c.Next()

		// 请求结束日志
		latency := time.Since(start)
		status := c.Writer.Status()

		slog.Info("access",
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("client_ip", clientIP),
			zap.Int("status", status),
			zap.Duration("latency", latency),
		)
	}
}
