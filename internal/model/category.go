package model

import (
	"time"

	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// カテゴリ情報
type Category struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Slug      string    `gorm:"type:varchar(50);not null;uniqueIndex"`
	Name      string    `gorm:"type:varchar(100);not null"`
	SortOrder int       `gorm:"not null;default:0;column:sort_order"`
	IsActive  bool      `gorm:"not null;default:true;column:is_active"`
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}
