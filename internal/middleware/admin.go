package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// Admin は管理者権限を持つユーザーのみアクセスを許可するミドルウェア
// Auth ミドルウェアの後に使用する必要がある
func Admin(userRepo repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.FromContext(c.Request.Context())

		// コンテキストからユーザー ID を取得
		userIDStr, exists := GetUserID(c)
		if !exists {
			log.Warn("コンテキストに user_id がありません")
			abortWithForbidden(c)
			return
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			log.Warn("user_id の形式が不正です", "user_id", userIDStr)
			abortWithForbidden(c)
			return
		}

		// ユーザー情報を取得
		user, err := userRepo.FindByID(c.Request.Context(), userID)
		if err != nil {
			log.Warn("ユーザーの取得に失敗しました", "user_id", userID, "error", err)
			abortWithForbidden(c)
			return
		}

		// 管理者権限をチェック
		if !user.Role.IsAdmin() {
			log.Warn("ユーザーに管理者権限がありません", "user_id", userID, "role", user.Role)
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
