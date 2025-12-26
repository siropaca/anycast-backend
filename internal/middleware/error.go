package middleware

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/logger"
)

// エラーをログに記録するミドルウェア
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// エラーがあれば処理
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			log := logger.FromContext(c.Request.Context())

			if appErr, ok := err.(*apperror.AppError); ok {
				log.Error("request error",
					slog.String("code", appErr.Code),
					slog.String("message", appErr.Message),
					slog.Any("underlying", appErr.Err),
				)
			} else {
				log.Error("unexpected error", slog.Any("error", err))
			}
		}
	}
}
