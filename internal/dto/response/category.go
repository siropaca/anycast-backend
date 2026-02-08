package response

import "github.com/siropaca/anycast-backend/internal/pkg/uuid"

// カテゴリ情報のレスポンス
type CategoryResponse struct {
	ID           uuid.UUID        `json:"id" validate:"required"`
	Slug         string           `json:"slug" validate:"required"`
	Name         string           `json:"name" validate:"required"`
	Image        *ArtworkResponse `json:"image"`
	ChannelCount int              `json:"channelCount" validate:"required"`
	EpisodeCount int              `json:"episodeCount" validate:"required"`
	SortOrder    int              `json:"sortOrder" validate:"required"`
	IsActive     bool             `json:"isActive" validate:"required"`
}

// カテゴリ単体のレスポンス
type CategoryDataResponse struct {
	Data CategoryResponse `json:"data" validate:"required"`
}

// カテゴリ一覧のレスポンス
type CategoryListResponse struct {
	Data []CategoryResponse `json:"data" validate:"required"`
}
