package response

import (
	"time"

	"github.com/google/uuid"
)

// チャンネル情報のレスポンス
type ChannelResponse struct {
	ID           uuid.UUID        `json:"id" validate:"required"`
	Name         string           `json:"name" validate:"required"`
	Description  string           `json:"description" validate:"required"`
	ScriptPrompt string           `json:"scriptPrompt" validate:"required"`
	Category     CategoryResponse `json:"category" validate:"required"`
	Artwork      *ArtworkResponse `json:"artwork"`
	PublishedAt  *time.Time       `json:"publishedAt"`
	CreatedAt    time.Time        `json:"createdAt" validate:"required"`
	UpdatedAt    time.Time        `json:"updatedAt" validate:"required"`
}

// チャンネル一覧（ページネーション付き）のレスポンス
type ChannelListWithPaginationResponse struct {
	Data       []ChannelResponse  `json:"data" validate:"required"`
	Pagination PaginationResponse `json:"pagination" validate:"required"`
}

// チャンネル単体のレスポンス
type ChannelDataResponse struct {
	Data ChannelResponse `json:"data" validate:"required"`
}
