package model

import (
	"time"

	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// Episode はエピソード情報を表す
type Episode struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ChannelID     uuid.UUID  `gorm:"type:uuid;not null;column:channel_id"`
	Title         string     `gorm:"type:varchar(255);not null"`
	Description   string     `gorm:"type:text;not null"`
	VoiceStyle    string     `gorm:"type:text;not null;default:'';column:voice_style"`
	ArtworkID     *uuid.UUID `gorm:"type:uuid;column:artwork_id"`
	BgmID         *uuid.UUID `gorm:"type:uuid;column:bgm_id"`
	SystemBgmID   *uuid.UUID `gorm:"type:uuid;column:system_bgm_id"`
	VoiceAudioID  *uuid.UUID `gorm:"type:uuid;column:voice_audio_id"`
	FullAudioID   *uuid.UUID `gorm:"type:uuid;column:full_audio_id"`
	PlayCount     int        `gorm:"not null;default:0;column:play_count"`
	PublishedAt   *time.Time `gorm:"column:published_at"`
	CreatedAt     time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt     time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP"`

	// リレーション
	Channel    Channel    `gorm:"foreignKey:ChannelID"`
	Artwork    *Image     `gorm:"foreignKey:ArtworkID"`
	Bgm        *Bgm       `gorm:"foreignKey:BgmID"`
	SystemBgm  *SystemBgm `gorm:"foreignKey:SystemBgmID"`
	VoiceAudio *Audio     `gorm:"foreignKey:VoiceAudioID"`
	FullAudio  *Audio     `gorm:"foreignKey:FullAudioID"`
}
