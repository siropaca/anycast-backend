package response

import (
	"time"

	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// 再生履歴内のエピソード情報のレスポンス
type PlaybackHistoryEpisodeResponse struct {
	ID          uuid.UUID                       `json:"id" validate:"required"`
	Title       string                          `json:"title" validate:"required"`
	Description string                          `json:"description" validate:"required"`
	FullAudio   *AudioResponse                  `json:"fullAudio" extensions:"x-nullable"`
	Channel     PlaybackHistoryChannelResponse  `json:"channel" validate:"required"`
	PublishedAt *time.Time                      `json:"publishedAt" extensions:"x-nullable"`
}

// 再生履歴内のチャンネル情報のレスポンス
type PlaybackHistoryChannelResponse struct {
	ID      uuid.UUID        `json:"id" validate:"required"`
	Name    string           `json:"name" validate:"required"`
	Artwork *ArtworkResponse `json:"artwork" extensions:"x-nullable"`
}

// 再生履歴アイテムのレスポンス
type PlaybackHistoryItemResponse struct {
	Episode    PlaybackHistoryEpisodeResponse `json:"episode" validate:"required"`
	ProgressMs int                            `json:"progressMs" validate:"required"`
	Completed  bool                           `json:"completed" validate:"required"`
	PlayedAt   time.Time                      `json:"playedAt" validate:"required"`
}

// 再生履歴一覧（ページネーション付き）のレスポンス
type PlaybackHistoryListWithPaginationResponse struct {
	Data       []PlaybackHistoryItemResponse `json:"data" validate:"required"`
	Pagination PaginationResponse            `json:"pagination" validate:"required"`
}

// 再生位置更新のレスポンス
type PlaybackResponse struct {
	EpisodeID  uuid.UUID `json:"episodeId" validate:"required"`
	ProgressMs int       `json:"progressMs" validate:"required"`
	Completed  bool      `json:"completed" validate:"required"`
	PlayedAt   time.Time `json:"playedAt" validate:"required"`
}

// 再生位置更新のラッパーレスポンス
type PlaybackDataResponse struct {
	Data PlaybackResponse `json:"data" validate:"required"`
}
