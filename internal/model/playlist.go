package model

import (
	"time"

	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// Playlist は再生リスト情報を表す
type Playlist struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;column:user_id"`
	Name        string    `gorm:"type:varchar(100);not null"`
	Description string    `gorm:"type:text;not null;default:''"`
	IsDefault   bool      `gorm:"not null;default:false;column:is_default"`
	CreatedAt   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`

	// リレーション
	Items []PlaylistItem `gorm:"foreignKey:PlaylistID"`
}

// PlaylistItem は再生リストアイテム情報を表す
type PlaylistItem struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PlaylistID uuid.UUID `gorm:"type:uuid;not null;column:playlist_id"`
	EpisodeID  uuid.UUID `gorm:"type:uuid;not null;column:episode_id"`
	Position   int       `gorm:"not null"`
	AddedAt    time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;column:added_at"`

	// リレーション
	Episode Episode `gorm:"foreignKey:EpisodeID"`
}
