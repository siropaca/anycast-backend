package response

import (
	"time"

	"github.com/google/uuid"
)

// ユーザー情報のレスポンス
type UserResponse struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	Username    string    `json:"username"`
	DisplayName string    `json:"displayName"`
	AvatarURL   *string   `json:"avatarUrl"`
}

// アバター情報
type AvatarResponse struct {
	ID  uuid.UUID `json:"id"`
	URL string    `json:"url"`
}

// 現在のユーザー情報（GET /auth/me 用）
type MeResponse struct {
	ID             uuid.UUID       `json:"id"`
	Email          string          `json:"email"`
	Username       string          `json:"username"`
	DisplayName    string          `json:"displayName"`
	Avatar         *AvatarResponse `json:"avatar"`
	HasPassword    bool            `json:"hasPassword"`
	OAuthProviders []string        `json:"oauthProviders"`
	CreatedAt      time.Time       `json:"createdAt"`
}

// 現在のユーザー情報のレスポンス（data ラッパー）
type MeDataResponse struct {
	Data MeResponse `json:"data"`
}

// ユーザー単体のレスポンス
type UserDataResponse struct {
	Data UserResponse `json:"data"`
}

// 認証成功時のレスポンス
type AuthResponse struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token"`
}

// 認証成功時のレスポンス（data ラッパー）
type AuthDataResponse struct {
	Data AuthResponse `json:"data"`
}
