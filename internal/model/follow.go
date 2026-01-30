package model

import (
	"time"

	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// Follow はユーザー間のフォロー関係を表す
type Follow struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID       uuid.UUID `gorm:"type:uuid;not null"`
	TargetUserID uuid.UUID `gorm:"type:uuid;not null;column:target_user_id"`
	CreatedAt    time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`

	TargetUser User `gorm:"foreignKey:TargetUserID"`
}
