package response

import (
	"time"

	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// BGM 一覧（ページネーション付き）のレスポンス
type BgmListWithPaginationResponse struct {
	Data       []BgmWithEpisodesResponse `json:"data" validate:"required"`
	Pagination PaginationResponse        `json:"pagination" validate:"required"`
}

// BGM 単体のレスポンス
type BgmDataResponse struct {
	Data BgmWithEpisodesResponse `json:"data" validate:"required"`
}

// エピソード・チャンネル情報付き BGM のレスポンス
type BgmWithEpisodesResponse struct {
	ID        uuid.UUID            `json:"id" validate:"required"`
	Name      string               `json:"name" validate:"required"`
	IsSystem  bool                 `json:"isSystem" validate:"required"`
	Audio     BgmAudioResponse     `json:"audio" validate:"required"`
	Episodes  []BgmEpisodeResponse `json:"episodes" validate:"required"`
	Channels  []BgmChannelResponse `json:"channels" validate:"required"`
	CreatedAt time.Time            `json:"createdAt" validate:"required"`
	UpdatedAt time.Time            `json:"updatedAt" validate:"required"`
}

// BGM のレスポンス
type BgmResponse struct {
	ID       uuid.UUID        `json:"id" validate:"required"`
	Name     string           `json:"name" validate:"required"`
	IsSystem bool             `json:"isSystem" validate:"required"`
	Audio    BgmAudioResponse `json:"audio" validate:"required"`
	CreatedAt time.Time        `json:"createdAt" validate:"required"`
	UpdatedAt time.Time        `json:"updatedAt" validate:"required"`
}

// BGM に紐づくエピソード情報のレスポンス
type BgmEpisodeResponse struct {
	ID      uuid.UUID                 `json:"id" validate:"required"`
	Title   string                    `json:"title" validate:"required"`
	Channel BgmEpisodeChannelResponse `json:"channel" validate:"required"`
}

// BGM エピソードに紐づくチャンネル情報のレスポンス
type BgmEpisodeChannelResponse struct {
	ID   uuid.UUID `json:"id" validate:"required"`
	Name string    `json:"name" validate:"required"`
}

// BGM に紐づくチャンネル情報のレスポンス
type BgmChannelResponse struct {
	ID   uuid.UUID `json:"id" validate:"required"`
	Name string    `json:"name" validate:"required"`
}

// BGM に紐づく音声情報のレスポンス
type BgmAudioResponse struct {
	ID         uuid.UUID `json:"id" validate:"required"`
	URL        string    `json:"url" validate:"required"`
	DurationMs int       `json:"durationMs" validate:"required"`
}
