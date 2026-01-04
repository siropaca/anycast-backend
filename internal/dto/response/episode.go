package response

import (
	"time"

	"github.com/google/uuid"
)

// エピソード情報のレスポンス
type EpisodeResponse struct {
	ID           uuid.UUID        `json:"id" validate:"required"`
	Title        string           `json:"title" validate:"required"`
	Description  *string          `json:"description"`
	ScriptPrompt *string          `json:"scriptPrompt,omitempty"`
	Artwork      *ArtworkResponse `json:"artwork"`
	FullAudio    *AudioResponse   `json:"fullAudio"`
	PublishedAt  *time.Time       `json:"publishedAt"`
	CreatedAt    time.Time        `json:"createdAt" validate:"required"`
	UpdatedAt    time.Time        `json:"updatedAt" validate:"required"`
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
