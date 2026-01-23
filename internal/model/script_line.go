package model

import (
	"time"

	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// ScriptLine は台本行情報を表す
type ScriptLine struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EpisodeID uuid.UUID `gorm:"type:uuid;not null;column:episode_id"`
	LineOrder int       `gorm:"not null;column:line_order"`
	SpeakerID uuid.UUID `gorm:"type:uuid;not null;column:speaker_id"`
	Text      string    `gorm:"type:text;not null"`
	Emotion   *string   `gorm:"type:text"`
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`

	// リレーション
	Episode Episode   `gorm:"foreignKey:EpisodeID"`
	Speaker Character `gorm:"foreignKey:SpeakerID"`
}

// TableName はテーブル名を返す
func (ScriptLine) TableName() string {
	return "script_lines"
}
