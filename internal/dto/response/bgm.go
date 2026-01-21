package response

import (
	"time"

	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// BGM 一覧（ページネーション付き）のレスポンス
type BgmListWithPaginationResponse struct {
	Data       []BgmResponse      `json:"data" validate:"required"`
	Pagination PaginationResponse `json:"pagination" validate:"required"`
}

// BGM 単体のレスポンス
type BgmDataResponse struct {
	Data BgmResponse `json:"data" validate:"required"`
}

// BGM のレスポンス
type BgmResponse struct {
	ID        uuid.UUID        `json:"id" validate:"required"`
	Name      string           `json:"name" validate:"required"`
	IsDefault bool             `json:"isDefault" validate:"required"`
	Audio     BgmAudioResponse `json:"audio" validate:"required"`
	CreatedAt time.Time        `json:"createdAt" validate:"required"`
	UpdatedAt time.Time        `json:"updatedAt" validate:"required"`
}

// BGM に紐づく音声情報のレスポンス
type BgmAudioResponse struct {
	ID         uuid.UUID `json:"id" validate:"required"`
	URL        string    `json:"url" validate:"required"`
	DurationMs int       `json:"durationMs" validate:"required"`
}
