package model

import (
	"time"

	"github.com/google/uuid"
)

type Voice struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Provider        string    `gorm:"type:varchar(50);not null" json:"provider"`
	ProviderVoiceID string    `gorm:"type:varchar(100);not null;column:provider_voice_id" json:"providerVoiceId"`
	Name            string    `gorm:"type:varchar(100);not null" json:"name"`
	Gender          string    `gorm:"type:varchar(20);not null" json:"gender"`
	IsActive        bool      `gorm:"not null;default:true" json:"isActive"`
	CreatedAt       time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"-"`
	UpdatedAt       time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"-"`
}
