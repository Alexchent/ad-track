package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP请求总数计数器
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "HTTP请求总数",
		},
		[]string{"method", "path", "status"},
	)

	// HTTP请求延迟直方图
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP请求延迟（秒）",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path"},
	)

	// 活跃请求数
	httpRequestsInFlight = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "当前正在处理的HTTP请求数",
		},
		[]string{"method", "path"},
	)
)

// PrometheusMetrics Prometheus指标中间件
func PrometheusMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		method := c.Request.Method

		// 如果没有匹配到路由，使用实际路径
		if path == "" {
			path = c.Request.URL.Path
		}

		// 增加活跃请求计数
		httpRequestsInFlight.WithLabelValues(method, path).Inc()
		defer httpRequestsInFlight.WithLabelValues(method, path).Dec()

		// 处理请求
		c.Next()

		// 记录请求指标
		status := strconv.Itoa(c.Writer.Status())
		duration := time.Since(start).Seconds()

		httpRequestsTotal.WithLabelValues(method, path, status).Inc()
		httpRequestDuration.WithLabelValues(method, path).Observe(duration)
	}
}
