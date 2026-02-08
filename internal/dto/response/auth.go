package response

import (
	"time"

	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// ユーザー情報のレスポンス
type UserResponse struct {
	ID          uuid.UUID `json:"id" validate:"required"`
	Email       string    `json:"email" validate:"required"`
	Username    string    `json:"username" validate:"required"`
	DisplayName string    `json:"displayName" validate:"required"`
	Role        string    `json:"role" validate:"required"`
	AvatarURL   *string   `json:"avatarUrl" extensions:"x-nullable"`
}

// アバター情報
type AvatarResponse struct {
	ID  uuid.UUID `json:"id" validate:"required"`
	URL string    `json:"url" validate:"required"`
}

// 現在のユーザー情報（GET /auth/me 用）
type MeResponse struct {
	ID             uuid.UUID       `json:"id" validate:"required"`
	Email          string          `json:"email" validate:"required"`
	Username       string          `json:"username" validate:"required"`
	DisplayName    string          `json:"displayName" validate:"required"`
	Bio            string          `json:"bio" validate:"required"`
	Role           string          `json:"role" validate:"required"`
	Avatar         *AvatarResponse `json:"avatar" extensions:"x-nullable"`
	HeaderImage    *AvatarResponse `json:"headerImage" extensions:"x-nullable"`
	UserPrompt     string          `json:"userPrompt" validate:"required"`
	HasPassword    bool            `json:"hasPassword" validate:"required"`
	OAuthProviders []string        `json:"oauthProviders" validate:"required"`
	CreatedAt      time.Time       `json:"createdAt" validate:"required"`
}

// 現在のユーザー情報のレスポンス（data ラッパー）
type MeDataResponse struct {
	Data MeResponse `json:"data" validate:"required"`
}

// ユーザー単体のレスポンス
type UserDataResponse struct {
	Data UserResponse `json:"data" validate:"required"`
}

// 認証成功時のレスポンス
type AuthResponse struct {
	User         UserResponse `json:"user" validate:"required"`
	AccessToken  string       `json:"accessToken" validate:"required"`
	RefreshToken string       `json:"refreshToken" validate:"required"`
}

// 認証成功時のレスポンス（data ラッパー）
type AuthDataResponse struct {
	Data AuthResponse `json:"data" validate:"required"`
}

// トークンリフレッシュ成功時のレスポンス
type TokenRefreshResponse struct {
	AccessToken  string `json:"accessToken" validate:"required"`
	RefreshToken string `json:"refreshToken" validate:"required"`
}

// トークンリフレッシュ成功時のレスポンス（data ラッパー）
type TokenRefreshDataResponse struct {
	Data TokenRefreshResponse `json:"data" validate:"required"`
}
