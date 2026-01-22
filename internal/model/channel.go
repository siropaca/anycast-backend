package model

import (
	"time"

	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// チャンネル情報
type Channel struct {
	ID                 uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID             uuid.UUID  `gorm:"type:uuid;not null;column:user_id"`
	Name               string     `gorm:"type:varchar(255);not null"`
	Description        string     `gorm:"type:text;not null"`
	UserPrompt         string     `gorm:"type:text;not null;column:user_prompt"`
	CategoryID         uuid.UUID  `gorm:"type:uuid;not null;column:category_id"`
	ArtworkID          *uuid.UUID `gorm:"type:uuid;column:artwork_id"`
	DefaultBgmID       *uuid.UUID `gorm:"type:uuid;column:default_bgm_id"`
	DefaultSystemBgmID *uuid.UUID `gorm:"type:uuid;column:default_system_bgm_id"`
	PublishedAt        *time.Time `gorm:"column:published_at"`
	CreatedAt          time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt          time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP"`

	// リレーション
	Category          Category           `gorm:"foreignKey:CategoryID"`
	Artwork           *Image             `gorm:"foreignKey:ArtworkID"`
	DefaultBgm        *Bgm               `gorm:"foreignKey:DefaultBgmID"`
	DefaultSystemBgm  *SystemBgm         `gorm:"foreignKey:DefaultSystemBgmID"`
	ChannelCharacters []ChannelCharacter `gorm:"foreignKey:ChannelID"`
}
