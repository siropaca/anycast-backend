package model

import (
	"time"

	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// Contact はお問い合わせ情報を表す
type Contact struct {
	ID        uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    *uuid.UUID      `gorm:"type:uuid;column:user_id"`
	Category  ContactCategory `gorm:"type:contact_category;not null"`
	Email     string          `gorm:"type:varchar(255);not null"`
	Name      string          `gorm:"type:varchar(100);not null"`
	Content   string          `gorm:"type:text;not null"`
	UserAgent *string         `gorm:"type:varchar(1024);column:user_agent"`
	CreatedAt time.Time       `gorm:"not null;default:CURRENT_TIMESTAMP"`

	// リレーション
	User *User `gorm:"foreignKey:UserID"`
}
