package model

import (
	"time"

	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// FavoriteVoice はユーザーのボイスお気に入り登録を表す
type FavoriteVoice struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID `gorm:"type:uuid;not null"`
	VoiceID   uuid.UUID `gorm:"type:uuid;not null;column:voice_id"`
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}
