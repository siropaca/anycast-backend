package model

import (
	"time"

	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// PlaybackHistory は再生履歴情報を表す
type PlaybackHistory struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;column:user_id"`
	EpisodeID  uuid.UUID `gorm:"type:uuid;not null;column:episode_id"`
	ProgressMs int       `gorm:"not null;default:0;column:progress_ms"`
	Completed  bool      `gorm:"not null;default:false"`
	PlayedAt   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;column:played_at"`
	CreatedAt  time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt  time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`

	// リレーション
	Episode Episode `gorm:"foreignKey:EpisodeID"`
}
