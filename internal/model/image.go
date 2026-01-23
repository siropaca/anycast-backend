package model

import (
	"time"

	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// Image は画像ファイル情報を表す
type Image struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	MimeType  string    `gorm:"type:varchar(100);not null;column:mime_type"`
	Path      string    `gorm:"type:varchar(1024);not null"`
	Filename  string    `gorm:"type:varchar(255);not null"`
	FileSize  int       `gorm:"not null;column:file_size"`
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}
