package response

import (
	"time"

	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// 話者情報のレスポンス
type SpeakerResponse struct {
	ID      uuid.UUID              `json:"id" validate:"required"`
	Name    string                 `json:"name" validate:"required"`
	Persona string                 `json:"persona" validate:"required"`
	Voice   CharacterVoiceResponse `json:"voice" validate:"required"`
}

// 台本行情報のレスポンス
type ScriptLineResponse struct {
	ID        uuid.UUID       `json:"id" validate:"required"`
	LineOrder int             `json:"lineOrder" validate:"required"`
	Speaker   SpeakerResponse `json:"speaker" validate:"required"`
	Text      string          `json:"text" validate:"required"`
	Emotion   *string         `json:"emotion,omitempty"`
	CreatedAt time.Time       `json:"createdAt" validate:"required"`
	UpdatedAt time.Time       `json:"updatedAt" validate:"required"`
}

// 台本行一覧のレスポンス
type ScriptLineListResponse struct {
	Data []ScriptLineResponse `json:"data" validate:"required"`
}
