package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/siropaca/anycast-backend/internal/pkg/logger"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// Logger はリクエストログを出力するミドルウェア
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		requestID := uuid.New().String()

		// リクエスト固有の Logger を作成
		log := logger.Default().With(
			slog.String("request_id", requestID),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
		)

		// Context に Logger を設定
		ctx := logger.WithContext(c.Request.Context(), log)
		c.Request = c.Request.WithContext(ctx)

		// ヘッダーにリクエスト ID を設定
		c.Header("X-Request-ID", requestID)

		c.Next()

		// レスポンスログ
		duration := time.Since(start)
		log.Debug("request completed",
			slog.Int("status", c.Writer.Status()),
			slog.Duration("duration", duration),
		)
	}
}
