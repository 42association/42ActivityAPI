package logging

import (
	"log/slog"
	"time"
	"github.com/gin-gonic/gin"
)

var Logger *slog.Logger

func GinLogger(logger *slog.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		fields := []any{
			slog.String("time", param.TimeStamp.Format(time.RFC3339)),
			slog.Int("status", param.StatusCode),
			slog.String("latency", param.Latency.String()),
			slog.String("client_ip", param.ClientIP),
			slog.String("method", param.Method),
			slog.String("path", param.Path),
			slog.String("error", param.ErrorMessage),
		}

		switch {
		case param.StatusCode >= 200 && param.StatusCode < 300:
			logger.Info("gin-request", fields...)
		case param.StatusCode >= 400 && param.StatusCode < 500:
			logger.Warn("gin-request", fields...)
		case param.StatusCode >= 500:
			logger.Error("gin-request", fields...)
		default:
			logger.Info("gin-request", fields...)
		}
		return ""
	})
}
