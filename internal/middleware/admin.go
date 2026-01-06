package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// 管理者権限を持つユーザーのみアクセスを許可するミドルウェア
// Auth ミドルウェアの後に使用する必要がある
func Admin(userRepo repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.FromContext(c.Request.Context())

		// コンテキストからユーザー ID を取得
		userIDStr, exists := GetUserID(c)
		if !exists {
			log.Warn("user_id not found in context")
			abortWithForbidden(c)
			return
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			log.Warn("invalid user_id format", "user_id", userIDStr)
			abortWithForbidden(c)
			return
		}

		// ユーザー情報を取得
		user, err := userRepo.FindByID(c.Request.Context(), userID)
		if err != nil {
			log.Warn("failed to find user", "user_id", userID, "error", err)
			abortWithForbidden(c)
			return
		}

		// 管理者権限をチェック
		if !user.Role.IsAdmin() {
			log.Warn("user does not have admin role", "user_id", userID, "role", user.Role)
			abortWithForbidden(c)
			return
		}

		c.Next()
	}
}

// 403 Forbidden エラーレスポンスを返して処理を中断する
func abortWithForbidden(c *gin.Context) {
	c.AbortWithStatusJSON(apperror.ErrForbidden.HTTPStatus, gin.H{
		"error": gin.H{
			"code":    apperror.ErrForbidden.Code,
			"message": apperror.ErrForbidden.Message,
		},
	})
}
