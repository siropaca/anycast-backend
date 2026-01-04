package response

import (
	"time"

	"github.com/google/uuid"
)

// 話者情報のレスポンス
type SpeakerResponse struct {
	ID   uuid.UUID `json:"id" validate:"required"`
	Name string    `json:"name" validate:"required"`
}

// 効果音情報のレスポンス
type SfxResponse struct {
	ID   uuid.UUID `json:"id" validate:"required"`
	Name string    `json:"name" validate:"required"`
}

// 台本行情報のレスポンス
type ScriptLineResponse struct {
	ID         uuid.UUID        `json:"id" validate:"required"`
	LineOrder  int              `json:"lineOrder" validate:"required"`
	LineType   string           `json:"lineType" validate:"required"`
	Speaker    *SpeakerResponse `json:"speaker,omitempty"`
	Text       *string          `json:"text,omitempty"`
	Emotion    *string          `json:"emotion,omitempty"`
	DurationMs *int             `json:"durationMs,omitempty"`
	Sfx        *SfxResponse     `json:"sfx,omitempty"`
	Volume     *float64         `json:"volume,omitempty"`
	Audio      *AudioResponse   `json:"audio,omitempty"`
	CreatedAt  time.Time        `json:"createdAt" validate:"required"`
	UpdatedAt  time.Time        `json:"updatedAt" validate:"required"`
}

// 台本行一覧のレスポンス
type ScriptLineListResponse struct {
	Data []ScriptLineResponse `json:"data" validate:"required"`
}
