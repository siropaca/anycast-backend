package response

import (
	"time"

	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// キャラクター一覧（ページネーション付き）のレスポンス
type CharacterListWithPaginationResponse struct {
	Data       []CharacterWithChannelsResponse `json:"data" validate:"required"`
	Pagination PaginationResponse              `json:"pagination" validate:"required"`
}

// キャラクター単体のレスポンス
type CharacterDataResponse struct {
	Data CharacterWithChannelsResponse `json:"data" validate:"required"`
}

// チャンネル情報付きキャラクターのレスポンス
type CharacterWithChannelsResponse struct {
	ID        uuid.UUID                  `json:"id" validate:"required"`
	Name      string                     `json:"name" validate:"required"`
	Persona   string                     `json:"persona" validate:"required"`
	Avatar    *AvatarResponse            `json:"avatar" extensions:"x-nullable"`
	Voice     CharacterVoiceResponse     `json:"voice" validate:"required"`
	Channels  []CharacterChannelResponse `json:"channels" validate:"required"`
	CreatedAt time.Time                  `json:"createdAt" validate:"required"`
	UpdatedAt time.Time                  `json:"updatedAt" validate:"required"`
}

// キャラクターに紐づくチャンネル情報のレスポンス
type CharacterChannelResponse struct {
	ID   uuid.UUID `json:"id" validate:"required"`
	Name string    `json:"name" validate:"required"`
}
