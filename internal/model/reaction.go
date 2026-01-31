package model

import (
	"time"

	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// Reaction はエピソードへのリアクションを表す
type Reaction struct {
	ID           uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID       uuid.UUID    `gorm:"type:uuid;not null"`
	EpisodeID    uuid.UUID    `gorm:"type:uuid;not null;column:episode_id"`
	ReactionType ReactionType `gorm:"type:reaction_type;not null;column:reaction_type"`
	CreatedAt    time.Time    `gorm:"not null;default:CURRENT_TIMESTAMP"`

	Episode Episode `gorm:"foreignKey:EpisodeID"`
}
