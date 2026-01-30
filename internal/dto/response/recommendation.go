package response

import (
	"time"

	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// おすすめチャンネル情報のレスポンス
type RecommendedChannelResponse struct {
	ID              uuid.UUID        `json:"id" validate:"required"`
	Name            string           `json:"name" validate:"required"`
	Description     string           `json:"description" validate:"required"`
	Artwork         *ArtworkResponse `json:"artwork" extensions:"x-nullable"`
	Category        CategoryResponse `json:"category" validate:"required"`
	EpisodeCount    int              `json:"episodeCount" validate:"required"`
	TotalPlayCount  int              `json:"totalPlayCount" validate:"required"`
	LatestEpisodeAt *time.Time       `json:"latestEpisodeAt" extensions:"x-nullable"`
}

// おすすめチャンネル一覧（ページネーション付き）のレスポンス
type RecommendedChannelListResponse struct {
	Data       []RecommendedChannelResponse `json:"data" validate:"required"`
	Pagination PaginationResponse           `json:"pagination" validate:"required"`
}
