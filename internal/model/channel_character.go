package model

import (
	"time"

	"github.com/google/uuid"
)

// チャンネルとキャラクターの紐づけを管理する中間テーブル
type ChannelCharacter struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ChannelID   uuid.UUID `gorm:"type:uuid;not null;column:channel_id"`
	CharacterID uuid.UUID `gorm:"type:uuid;not null;column:character_id"`
	CreatedAt   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`

	// リレーション
	Character Character `gorm:"foreignKey:CharacterID"`
}
