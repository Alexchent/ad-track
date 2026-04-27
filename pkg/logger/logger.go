package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

const TraceIDKey = "traceID"
const UserIDKey = "user_id"

// TraceHandler 从日志 context 中自动补充 traceId 字段。
type TraceHandler struct {
	handler slog.Handler
}

func NewTraceHandler(handler slog.Handler) *TraceHandler {
	return &TraceHandler{handler: handler}
}

func (h *TraceHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *TraceHandler) Handle(ctx context.Context, record slog.Record) error {
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok && traceID != "" {
		record.AddAttrs(slog.String("traceId", traceID))
	}
	if userID, ok := ctx.Value(UserIDKey).(string); ok && userID != "" {
		record.AddAttrs(slog.String("user_id", userID))
	}
	return h.handler.Handle(ctx, record)
}

func (h *TraceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &TraceHandler{handler: h.handler.WithAttrs(attrs)}
}

func (h *TraceHandler) WithGroup(name string) slog.Handler {
	return &TraceHandler{handler: h.handler.WithGroup(name)}
}

type LogConfig struct {
	Filename string
	Encoding string `json:",optional"`
	Level    string `json:",optional"`
	MaxSize  int
	MaxAge   int
	Compress bool
}

func SetupLogger(conf *LogConfig) {
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
	logger := slog.New(NewTraceHandler(handler))
	slog.SetDefault(logger)
}
