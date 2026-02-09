package response

import (
	"time"

	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// チャンネル検索結果のレスポンス
type SearchChannelResponse struct {
	ID          uuid.UUID        `json:"id" validate:"required"`
	Name        string           `json:"name" validate:"required"`
	Description string           `json:"description" validate:"required"`
	Category    CategoryResponse `json:"category" validate:"required"`
	Artwork     *ArtworkResponse `json:"artwork" extensions:"x-nullable"`
	PublishedAt *time.Time       `json:"publishedAt" extensions:"x-nullable"`
	CreatedAt   time.Time        `json:"createdAt" validate:"required"`
	UpdatedAt   time.Time        `json:"updatedAt" validate:"required"`
}

// チャンネル検索結果一覧（ページネーション付き）のレスポンス
type SearchChannelListResponse struct {
	Data       []SearchChannelResponse `json:"data" validate:"required"`
	Pagination PaginationResponse      `json:"pagination" validate:"required"`
}

// エピソード検索結果のレスポンス
type SearchEpisodeResponse struct {
	ID          uuid.UUID                    `json:"id" validate:"required"`
	Title       string                       `json:"title" validate:"required"`
	Description string                       `json:"description" validate:"required"`
	Channel     SearchEpisodeChannelResponse `json:"channel" validate:"required"`
	PublishedAt *time.Time                   `json:"publishedAt" extensions:"x-nullable"`
	CreatedAt   time.Time                    `json:"createdAt" validate:"required"`
	UpdatedAt   time.Time                    `json:"updatedAt" validate:"required"`
}

// エピソード検索結果内のチャンネル情報のレスポンス
type SearchEpisodeChannelResponse struct {
	ID   uuid.UUID `json:"id" validate:"required"`
	Name string    `json:"name" validate:"required"`
}

// エピソード検索結果一覧（ページネーション付き）のレスポンス
type SearchEpisodeListResponse struct {
	Data       []SearchEpisodeResponse `json:"data" validate:"required"`
	Pagination PaginationResponse      `json:"pagination" validate:"required"`
}
