package response

import (
	"time"

	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// 孤児音声ファイル情報
type OrphanedAudioResponse struct {
	ID        uuid.UUID `json:"id" validate:"required"`
	URL       string    `json:"url" validate:"required"`
	Filename  string    `json:"filename" validate:"required"`
	FileSize  int       `json:"fileSize" validate:"required"`
	CreatedAt time.Time `json:"createdAt" validate:"required"`
}

// 孤児画像ファイル情報
type OrphanedImageResponse struct {
	ID        uuid.UUID `json:"id" validate:"required"`
	URL       string    `json:"url" validate:"required"`
	Filename  string    `json:"filename" validate:"required"`
	FileSize  int       `json:"fileSize" validate:"required"`
	CreatedAt time.Time `json:"createdAt" validate:"required"`
}

// 孤児メディアファイル削除レスポンス
type CleanupOrphanedMediaResponse struct {
	DryRun            bool                    `json:"dryRun" validate:"required"`
	OrphanedAudios    []OrphanedAudioResponse `json:"orphanedAudios" validate:"required"`
	OrphanedImages    []OrphanedImageResponse `json:"orphanedImages" validate:"required"`
	DeletedAudioCount int                     `json:"deletedAudioCount" validate:"required"`
	DeletedImageCount int                     `json:"deletedImageCount" validate:"required"`
	FailedAudioCount  int                     `json:"failedAudioCount" validate:"required"`
	FailedImageCount  int                     `json:"failedImageCount" validate:"required"`
}
