package response

import (
	"time"

	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// エピソード情報のレスポンス
type EpisodeResponse struct {
	ID          uuid.UUID                `json:"id" validate:"required"`
	Owner       ChannelOwnerResponse     `json:"owner" validate:"required"`
	Title       string                   `json:"title" validate:"required"`
	Description string                   `json:"description" validate:"required"`
	Artwork     *ArtworkResponse         `json:"artwork" extensions:"x-nullable"`
	VoiceAudio  *AudioResponse           `json:"voiceAudio,omitempty" extensions:"x-nullable"`
	FullAudio   *AudioResponse           `json:"fullAudio" extensions:"x-nullable"`
	Bgm         *EpisodeBgmResponse      `json:"bgm" extensions:"x-nullable"`
	Playback    *EpisodePlaybackResponse `json:"playback" extensions:"x-nullable"`
	PlaylistIDs []uuid.UUID              `json:"playlistIds" extensions:"x-nullable"`
	PlayCount   int                      `json:"playCount" validate:"required"`
	PublishedAt *time.Time               `json:"publishedAt" extensions:"x-nullable"`
	CreatedAt   time.Time                `json:"createdAt" validate:"required"`
	UpdatedAt   time.Time                `json:"updatedAt" validate:"required"`
}

// エピソードの再生位置情報のレスポンス
type EpisodePlaybackResponse struct {
	ProgressMs int       `json:"progressMs" validate:"required"`
	Completed  bool      `json:"completed" validate:"required"`
	PlayedAt   time.Time `json:"playedAt" validate:"required"`
}

// エピソードに設定された BGM のレスポンス
type EpisodeBgmResponse struct {
	ID       uuid.UUID        `json:"id" validate:"required"`
	Name     string           `json:"name" validate:"required"`
	IsSystem bool             `json:"isSystem" validate:"required"`
	Audio    BgmAudioResponse `json:"audio" validate:"required"`
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
