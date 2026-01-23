package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"google.golang.org/api/idtoken"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
)

// CloudTasksAuth は Cloud Tasks からの OIDC トークンを検証するミドルウェア
//
// expectedAudience はワーカーエンドポイントの URL（Cloud Tasks が OIDC トークンを生成する際に指定する audience）
// expectedServiceAccount は Cloud Tasks に設定されたサービスアカウントのメールアドレス
func CloudTasksAuth(expectedAudience, expectedServiceAccount string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.FromContext(c.Request.Context())

		// Authorization ヘッダーからトークンを取得
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Warn("missing authorization header for cloud tasks")
			abortWithUnauthorized(c)
			return
		}

		// Bearer プレフィックスを確認
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			log.Warn("invalid authorization header format for cloud tasks")
			abortWithUnauthorized(c)
			return
		}

		tokenString := parts[1]

		// OIDC トークンを検証
		payload, err := validateOIDCToken(c.Request.Context(), tokenString, expectedAudience)
		if err != nil {
			log.Warn("invalid OIDC token", "error", err)
			abortWithUnauthorized(c)
			return
		}

		// サービスアカウントを確認
		email, ok := payload.Claims["email"].(string)
		if !ok || email != expectedServiceAccount {
			log.Warn("invalid service account", "expected", expectedServiceAccount, "got", email)
			abortWithUnauthorized(c)
			return
		}

		log.Debug("cloud tasks auth successful", "email", email)
		c.Next()
	}
}

// validateOIDCToken は Google OIDC トークンを検証する
func validateOIDCToken(ctx context.Context, tokenString, expectedAudience string) (*idtoken.Payload, error) {
	payload, err := idtoken.Validate(ctx, tokenString, expectedAudience)
	if err != nil {
		return nil, apperror.ErrUnauthorized.WithMessage("OIDC トークンの検証に失敗しました").WithError(err)
	}

	return payload, nil
}
