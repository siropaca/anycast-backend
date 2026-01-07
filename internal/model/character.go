package model

import (
	"time"

	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// キャラクター情報
type Character struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;column:user_id"`
	Name      string     `gorm:"type:varchar(255);not null"`
	Persona   string     `gorm:"type:text;not null"`
	AvatarID  *uuid.UUID `gorm:"type:uuid;column:avatar_id"`
	VoiceID   uuid.UUID  `gorm:"type:uuid;not null;column:voice_id"`
	CreatedAt time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP"`

	// リレーション
	Avatar            *Image             `gorm:"foreignKey:AvatarID"`
	Voice             Voice              `gorm:"foreignKey:VoiceID"`
	ChannelCharacters []ChannelCharacter `gorm:"foreignKey:CharacterID"`
}
