package model

import (
	"time"

	"github.com/google/uuid"
)

// ユーザー情報
type User struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email       string     `gorm:"type:varchar(255);not null;uniqueIndex"`
	Username    string     `gorm:"type:varchar(20);not null;uniqueIndex"`
	DisplayName string     `gorm:"type:varchar(20);not null;column:display_name"`
	AvatarID    *uuid.UUID `gorm:"type:uuid"`
	CreatedAt   time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

// パスワード認証情報
type Credential struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	PasswordHash string    `gorm:"type:varchar(255);not null;column:password_hash"`
	CreatedAt    time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt    time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

// OAuth 認証情報
type OAuthAccount struct {
	ID             uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID         uuid.UUID     `gorm:"type:uuid;not null;index"`
	Provider       OAuthProvider `gorm:"type:oauth_provider;not null"`
	ProviderUserID string        `gorm:"type:varchar(255);not null;column:provider_user_id"`
	AccessToken    *string       `gorm:"type:varchar(1024);column:access_token"`
	RefreshToken   *string       `gorm:"type:varchar(1024);column:refresh_token"`
	ExpiresAt      *time.Time    `gorm:"column:expires_at"`
	CreatedAt      time.Time     `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt      time.Time     `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

func (OAuthAccount) TableName() string {
	return "oauth_accounts"
}
