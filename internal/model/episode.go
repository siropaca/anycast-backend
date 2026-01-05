package model

import (
	"time"

	"github.com/google/uuid"
)

// エピソード情報
type Episode struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ChannelID   uuid.UUID  `gorm:"type:uuid;not null;column:channel_id"`
	Title       string     `gorm:"type:varchar(255);not null"`
	Description string     `gorm:"type:text;not null"`
	UserPrompt  *string    `gorm:"type:text;column:user_prompt"`
	ArtworkID   *uuid.UUID `gorm:"type:uuid;column:artwork_id"`
	BgmID       *uuid.UUID `gorm:"type:uuid;column:bgm_id"`
	FullAudioID *uuid.UUID `gorm:"type:uuid;column:full_audio_id"`
	PublishedAt *time.Time `gorm:"column:published_at"`
	CreatedAt   time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP"`

	// リレーション
	Channel   Channel `gorm:"foreignKey:ChannelID"`
	Artwork   *Image  `gorm:"foreignKey:ArtworkID"`
	Bgm       *Audio  `gorm:"foreignKey:BgmID"`
	FullAudio *Audio  `gorm:"foreignKey:FullAudioID"`
}
