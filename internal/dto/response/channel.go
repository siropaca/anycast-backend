package response

import (
	"time"

	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// チャンネル情報のレスポンス
type ChannelResponse struct {
	ID          uuid.UUID                  `json:"id" validate:"required"`
	Name        string                     `json:"name" validate:"required"`
	Description string                     `json:"description" validate:"required"`
	UserPrompt  string                     `json:"userPrompt" validate:"required"`
	Category    CategoryResponse           `json:"category" validate:"required"`
	Artwork     *ArtworkResponse           `json:"artwork" extensions:"x-nullable"`
	DefaultBgm  *ChannelDefaultBgmResponse `json:"defaultBgm" extensions:"x-nullable"`
	Characters  []CharacterResponse        `json:"characters" validate:"required"`
	Episodes    []EpisodeResponse          `json:"episodes" validate:"required"`
	PublishedAt *time.Time                 `json:"publishedAt" extensions:"x-nullable"`
	CreatedAt   time.Time                  `json:"createdAt" validate:"required"`
	UpdatedAt   time.Time                  `json:"updatedAt" validate:"required"`
}

// チャンネルのデフォルト BGM 情報のレスポンス
type ChannelDefaultBgmResponse struct {
	ID        uuid.UUID        `json:"id" validate:"required"`
	Name      string           `json:"name" validate:"required"`
	IsDefault bool             `json:"isDefault" validate:"required"` // true=システムBGM
	Audio     BgmAudioResponse `json:"audio" validate:"required"`
}

// キャラクター情報のレスポンス
type CharacterResponse struct {
	ID        uuid.UUID              `json:"id" validate:"required"`
	Name      string                 `json:"name" validate:"required"`
	Persona   string                 `json:"persona" validate:"required"`
	Voice     CharacterVoiceResponse `json:"voice" validate:"required"`
	CreatedAt time.Time              `json:"createdAt" validate:"required"`
	UpdatedAt time.Time              `json:"updatedAt" validate:"required"`
}

// キャラクターに紐づくボイス情報のレスポンス
type CharacterVoiceResponse struct {
	ID       uuid.UUID `json:"id" validate:"required"`
	Name     string    `json:"name" validate:"required"`
	Provider string    `json:"provider" validate:"required"`
	Gender   string    `json:"gender" validate:"required"`
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
