package model

import (
	"time"

	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// 効果音情報
type SoundEffect struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name        string    `gorm:"type:varchar(100);not null;uniqueIndex"`
	Description *string   `gorm:"type:text"`
	AudioID     uuid.UUID `gorm:"type:uuid;not null;column:audio_id"`
	CreatedAt   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`

	// リレーション
	Audio Audio `gorm:"foreignKey:AudioID"`
}
