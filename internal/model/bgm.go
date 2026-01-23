package model

import (
	"time"

	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// Bgm はユーザーが所有する BGM を表す
type Bgm struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;column:user_id"`
	AudioID   uuid.UUID `gorm:"type:uuid;not null;column:audio_id"`
	Name      string    `gorm:"type:varchar(255);not null"`
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`

	// リレーション
	Audio    Audio     `gorm:"foreignKey:AudioID"`
	Episodes []Episode `gorm:"foreignKey:BgmID"`
	Channels []Channel `gorm:"foreignKey:DefaultBgmID"`
}
