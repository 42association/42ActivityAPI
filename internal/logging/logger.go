package logging

import (
	"log/slog"
	"time"
	"github.com/gin-gonic/gin"
)

var Logger *slog.Logger

func GinLogger(logger *slog.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		Logger.Info("gin-request",
			slog.String("time", param.TimeStamp.Format(time.RFC3339)),
			slog.Int("status", param.StatusCode),
			slog.String("latency", param.Latency.String()),
			slog.String("client_ip", param.ClientIP),
			slog.String("method", param.Method),
			slog.String("path", param.Path),
			slog.String("error", param.ErrorMessage),
		)
		return ""
	})
}