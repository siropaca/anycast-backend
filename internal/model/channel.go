package model

import (
	"time"

	"github.com/google/uuid"
)

// カテゴリ情報
type Category struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Slug      string    `gorm:"type:varchar(50);not null;uniqueIndex"`
	Name      string    `gorm:"type:varchar(100);not null"`
	SortOrder int       `gorm:"not null;default:0;column:sort_order"`
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

// チャンネル情報
type Channel struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null;column:user_id"`
	Name         string     `gorm:"type:varchar(255);not null"`
	Description  string     `gorm:"type:text;not null"`
	ScriptPrompt string     `gorm:"type:text;not null;column:script_prompt"`
	CategoryID   uuid.UUID  `gorm:"type:uuid;not null;column:category_id"`
	ArtworkID    *uuid.UUID `gorm:"type:uuid;column:artwork_id"`
	PublishedAt  *time.Time `gorm:"column:published_at"`
	CreatedAt    time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt    time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP"`

	// リレーション
	Category Category `gorm:"foreignKey:CategoryID"`
	Artwork  *Image   `gorm:"foreignKey:ArtworkID"`
}
