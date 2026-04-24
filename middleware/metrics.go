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

	// 文件上传大小
	fileUploadSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "file_upload_size_bytes",
			Help:    "文件上传大小（字节）",
			Buckets: []float64{1024, 10 * 1024, 100 * 1024, 1024 * 1024, 10 * 1024 * 1024, 100 * 1024 * 1024},
		},
		[]string{"extension"},
	)

	// 任务提交计数
	taskSubmitTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "task_submit_total",
			Help: "任务提交总数",
		},
		[]string{"tp_id", "workflow_id"},
	)

	// 任务状态计数
	taskStatusTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "task_status_current",
			Help: "当前各状态的任务数",
		},
		[]string{"status"},
	)

	// 认证计数
	authTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_total",
			Help: "认证请求总数",
		},
		[]string{"type", "status"},
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

// RecordFileUpload 记录文件上传指标
func RecordFileUpload(extension string, size int64) {
	fileUploadSize.WithLabelValues(extension).Observe(float64(size))
}

// RecordTaskSubmit 记录任务提交指标
func RecordTaskSubmit(tpID, workflowID int) {
	taskSubmitTotal.WithLabelValues(
		strconv.Itoa(tpID),
		strconv.Itoa(workflowID),
	).Inc()
}

// UpdateTaskStatus 更新任务状态指标
func UpdateTaskStatus(status string, count float64) {
	taskStatusTotal.WithLabelValues(status).Set(count)
}

// RecordAuth 记录认证指标
func RecordAuth(authType string, success bool) {
	status := "success"
	if !success {
		status = "failed"
	}
	authTotal.WithLabelValues(authType, status).Inc()
}
