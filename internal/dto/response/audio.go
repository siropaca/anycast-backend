package response

import "github.com/siropaca/anycast-backend/internal/pkg/uuid"

// 音声ファイル情報のレスポンス
type AudioResponse struct {
	ID         uuid.UUID `json:"id" validate:"required"`
	URL        string    `json:"url" validate:"required"`
	MimeType   string    `json:"mimeType" validate:"required"`
	FileSize   int       `json:"fileSize" validate:"required"`
	DurationMs int       `json:"durationMs" validate:"required"`
}

// 音声生成レスポンス
type GenerateAudioResponse struct {
	Audio AudioResponse `json:"audio" validate:"required"`
}
