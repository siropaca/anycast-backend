package model

import (
	"time"

	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// APIKey はユーザーの API キーを表す
type APIKey struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID     uuid.UUID  `gorm:"type:uuid;not null;index"`
	Name       string     `gorm:"type:varchar(100);not null"`
	KeyHash    string     `gorm:"type:varchar(64);not null;uniqueIndex"`
	Prefix     string     `gorm:"type:varchar(20);not null"`
	LastUsedAt *time.Time `gorm:"column:last_used_at"`
	CreatedAt  time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt  time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP"`
}
