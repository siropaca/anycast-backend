package response

import (
	"time"

	"github.com/google/uuid"
)

// エピソード情報のレスポンス
type EpisodeResponse struct {
	ID          uuid.UUID        `json:"id" validate:"required"`
	Title       string           `json:"title" validate:"required"`
	Description string           `json:"description" validate:"required"`
	UserPrompt  *string          `json:"userPrompt,omitempty"`
	Artwork     *ArtworkResponse `json:"artwork" extensions:"x-nullable"`
	FullAudio   *AudioResponse   `json:"fullAudio" extensions:"x-nullable"`
	PublishedAt *time.Time       `json:"publishedAt" extensions:"x-nullable"`
	CreatedAt   time.Time        `json:"createdAt" validate:"required"`
	UpdatedAt   time.Time        `json:"updatedAt" validate:"required"`
}

// エピソード一覧（ページネーション付き）のレスポンス
type EpisodeListWithPaginationResponse struct {
	Data       []EpisodeResponse  `json:"data" validate:"required"`
	Pagination PaginationResponse `json:"pagination" validate:"required"`
}

// エピソード詳細のレスポンス
type EpisodeDataResponse struct {
	Data EpisodeResponse `json:"data" validate:"required"`
}
