package model

import (
	"time"

	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// SystemBgm はシステム BGM（管理者が提供）を表す
type SystemBgm struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AudioID   uuid.UUID `gorm:"type:uuid;not null;column:audio_id"`
	Name      string    `gorm:"type:varchar(255);not null"`
	SortOrder int       `gorm:"not null;default:0;column:sort_order"`
	IsActive  bool      `gorm:"not null;default:true;column:is_active"`
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`

	// リレーション
	Audio    Audio     `gorm:"foreignKey:AudioID"`
	Episodes []Episode `gorm:"foreignKey:SystemBgmID"`
	Channels []Channel `gorm:"foreignKey:DefaultSystemBgmID"`
}

// TableName はテーブル名を返す（マイグレーション後の新テーブル名）
func (SystemBgm) TableName() string {
	return "system_bgms"
}
