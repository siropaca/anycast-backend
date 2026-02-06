package response

import (
	"time"

	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// 公開ユーザーのチャンネル情報のレスポンス
type PublicUserChannelResponse struct {
	ID          uuid.UUID        `json:"id" validate:"required"`
	Name        string           `json:"name" validate:"required"`
	Description string           `json:"description" validate:"required"`
	Category    CategoryResponse `json:"category" validate:"required"`
	Artwork     *ArtworkResponse `json:"artwork" extensions:"x-nullable"`
	PublishedAt *time.Time       `json:"publishedAt" extensions:"x-nullable"`
	CreatedAt   time.Time        `json:"createdAt" validate:"required"`
	UpdatedAt   time.Time        `json:"updatedAt" validate:"required"`
}

// チャンネルオーナー情報のレスポンス
type ChannelOwnerResponse struct {
	ID          uuid.UUID       `json:"id" validate:"required"`
	Username    string          `json:"username" validate:"required"`
	DisplayName string          `json:"displayName" validate:"required"`
	Avatar      *AvatarResponse `json:"avatar" extensions:"x-nullable"`
}

// 公開ユーザー情報のレスポンス
type PublicUserResponse struct {
	ID          uuid.UUID                   `json:"id" validate:"required"`
	Username    string                      `json:"username" validate:"required"`
	DisplayName string                      `json:"displayName" validate:"required"`
	Avatar      *AvatarResponse             `json:"avatar" extensions:"x-nullable"`
	Channels    []PublicUserChannelResponse `json:"channels" validate:"required"`
	CreatedAt   time.Time                   `json:"createdAt" validate:"required"`
}

// 公開ユーザー情報のレスポンス（data ラッパー）
type PublicUserDataResponse struct {
	Data PublicUserResponse `json:"data" validate:"required"`
}
