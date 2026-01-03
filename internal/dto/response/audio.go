package response

import "github.com/google/uuid"

// 音声ファイル情報のレスポンス
type AudioResponse struct {
	ID         uuid.UUID `json:"id" validate:"required"`
	URL        string    `json:"url" validate:"required"`
	DurationMs int       `json:"durationMs" validate:"required"`
}
