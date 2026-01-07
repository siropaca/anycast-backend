package response

import "github.com/siropaca/anycast-backend/internal/pkg/uuid"

// アートワーク情報のレスポンス
type ArtworkResponse struct {
	ID  uuid.UUID `json:"id" validate:"required"`
	URL string    `json:"url" validate:"required"`
}

// 画像アップロードのレスポンス
type ImageUploadResponse struct {
	ID       uuid.UUID `json:"id" validate:"required"`
	MimeType string    `json:"mimeType" validate:"required"`
	URL      string    `json:"url" validate:"required"`
	Filename string    `json:"filename" validate:"required"`
	FileSize int       `json:"fileSize" validate:"required"`
}

// 画像アップロードのレスポンス（data ラッパー付き）
type ImageUploadDataResponse struct {
	Data ImageUploadResponse `json:"data" validate:"required"`
}
