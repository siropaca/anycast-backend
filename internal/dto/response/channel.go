package response

import (
	"time"

	"github.com/google/uuid"
)

// チャンネル情報のレスポンス
type ChannelResponse struct {
	ID           uuid.UUID        `json:"id"`
	Name         string           `json:"name"`
	Description  *string          `json:"description"`
	ScriptPrompt *string          `json:"scriptPrompt"`
	Category     CategoryResponse `json:"category"`
	Artwork      *ArtworkResponse `json:"artwork"`
	PublishedAt  *time.Time       `json:"publishedAt"`
	CreatedAt    time.Time        `json:"createdAt"`
	UpdatedAt    time.Time        `json:"updatedAt"`
}

// チャンネル一覧（ページネーション付き）のレスポンス
type ChannelListWithPaginationResponse struct {
	Data       []ChannelResponse  `json:"data"`
	Pagination PaginationResponse `json:"pagination"`
}
