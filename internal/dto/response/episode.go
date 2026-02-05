package response

import (
	"time"

	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// エピソード情報のレスポンス
type EpisodeResponse struct {
	ID            uuid.UUID           `json:"id" validate:"required"`
	Title         string              `json:"title" validate:"required"`
	Description   string              `json:"description" validate:"required"`
	VoiceStyle    string              `json:"voiceStyle" validate:"required"`
	Artwork       *ArtworkResponse    `json:"artwork" extensions:"x-nullable"`
	FullAudio     *AudioResponse      `json:"fullAudio" extensions:"x-nullable"`
	Bgm           *EpisodeBgmResponse `json:"bgm" extensions:"x-nullable"`
	AudioOutdated bool                `json:"audioOutdated" validate:"required"`
	PlayCount     int                 `json:"playCount" validate:"required"`
	PublishedAt   *time.Time          `json:"publishedAt" extensions:"x-nullable"`
	CreatedAt     time.Time           `json:"createdAt" validate:"required"`
	UpdatedAt     time.Time           `json:"updatedAt" validate:"required"`
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
