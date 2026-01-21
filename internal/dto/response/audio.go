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
	Data AudioResponse `json:"data" validate:"required"`
}

// 音声アップロードのレスポンス
type AudioUploadResponse struct {
	ID         uuid.UUID `json:"id" validate:"required"`
	MimeType   string    `json:"mimeType" validate:"required"`
	URL        string    `json:"url" validate:"required"`
	Filename   string    `json:"filename" validate:"required"`
	FileSize   int       `json:"fileSize" validate:"required"`
	DurationMs int       `json:"durationMs" validate:"required"`
}

// 音声アップロードのレスポンス（data ラッパー付き）
type AudioUploadDataResponse struct {
	Data AudioUploadResponse `json:"data" validate:"required"`
}
