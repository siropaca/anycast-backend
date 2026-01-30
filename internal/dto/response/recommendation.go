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

// おすすめエピソード情報のレスポンス
type RecommendedEpisodeResponse struct {
	ID                uuid.UUID                         `json:"id" validate:"required"`
	Title             string                            `json:"title" validate:"required"`
	Description       string                            `json:"description" validate:"required"`
	Artwork           *ArtworkResponse                  `json:"artwork" extensions:"x-nullable"`
	FullAudio         *AudioResponse                    `json:"fullAudio" extensions:"x-nullable"`
	PlayCount         int                               `json:"playCount" validate:"required"`
	PublishedAt       *time.Time                        `json:"publishedAt" extensions:"x-nullable"`
	Channel           RecommendedEpisodeChannelResponse `json:"channel" validate:"required"`
	PlaybackProgress  *PlaybackProgressResponse         `json:"playbackProgress" extensions:"x-nullable"`
	InDefaultPlaylist bool                              `json:"inDefaultPlaylist" validate:"required"`
}

// おすすめエピソード内のチャンネル情報のレスポンス
type RecommendedEpisodeChannelResponse struct {
	ID       uuid.UUID        `json:"id" validate:"required"`
	Name     string           `json:"name" validate:"required"`
	Artwork  *ArtworkResponse `json:"artwork" extensions:"x-nullable"`
	Category CategoryResponse `json:"category" validate:"required"`
}

// 再生進捗のレスポンス
type PlaybackProgressResponse struct {
	ProgressMs int  `json:"progressMs" validate:"required"`
	Completed  bool `json:"completed" validate:"required"`
}

// おすすめエピソード一覧（ページネーション付き）のレスポンス
type RecommendedEpisodeListResponse struct {
	Data       []RecommendedEpisodeResponse `json:"data" validate:"required"`
	Pagination PaginationResponse           `json:"pagination" validate:"required"`
}
