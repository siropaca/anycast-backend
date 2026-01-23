package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/pkg/jwt"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
)

// コンテキストキー
type contextKey string

const (
	// UserIDKey はコンテキストからユーザー ID を取得するためのキー
	UserIDKey contextKey = "user_id"
)

// Auth は Bearer Token 認証を行うミドルウェア
func Auth(tokenManager jwt.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.FromContext(c.Request.Context())

		// Authorization ヘッダーからトークンを取得
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Warn("missing authorization header")
			abortWithUnauthorized(c)
			return
		}

		// Bearer プレフィックスを確認
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			log.Warn("invalid authorization header format")
			abortWithUnauthorized(c)
			return
		}

		tokenString := parts[1]

		// JWT を検証
		claims, err := tokenManager.Validate(tokenString)
		if err != nil {
			log.Warn("invalid token", "error", err)
			abortWithUnauthorized(c)
			return
		}

		// ユーザー ID をコンテキストに設定
		c.Set(string(UserIDKey), claims.UserID)

		c.Next()
	}
}

// 401 Unauthorized エラーレスポンスを返して処理を中断する
func abortWithUnauthorized(c *gin.Context) {
	c.AbortWithStatusJSON(apperror.ErrUnauthorized.HTTPStatus, gin.H{
		"error": gin.H{
			"code":    apperror.ErrUnauthorized.Code,
			"message": apperror.ErrUnauthorized.Message,
		},
	})
}

// GetUserID はコンテキストからユーザー ID を取得する
func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get(string(UserIDKey))
	if !exists {
		return "", false
	}
	id, ok := userID.(string)
	return id, ok
}
