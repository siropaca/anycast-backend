package model

import (
	"time"

	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// Feedback はフィードバック情報を表す
type Feedback struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null;column:user_id"`
	Content      string     `gorm:"type:text;not null"`
	ScreenshotID *uuid.UUID `gorm:"type:uuid;column:screenshot_id"`
	PageURL      *string    `gorm:"type:varchar(2048);column:page_url"`
	UserAgent    *string    `gorm:"type:varchar(1024);column:user_agent"`
	CreatedAt    time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP"`

	// リレーション
	User       User   `gorm:"foreignKey:UserID"`
	Screenshot *Image `gorm:"foreignKey:ScreenshotID"`
}
