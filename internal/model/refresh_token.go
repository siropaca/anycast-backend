package model

import (
	"time"

	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// RefreshToken はリフレッシュトークンを表す
type RefreshToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	Token     string    `gorm:"type:varchar(255);not null;uniqueIndex"`
	ExpiresAt time.Time `gorm:"not null;column:expires_at"`
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}
