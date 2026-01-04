package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// 台本行情報
type ScriptLine struct {
	ID         uuid.UUID        `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EpisodeID  uuid.UUID        `gorm:"type:uuid;not null;column:episode_id"`
	LineOrder  int              `gorm:"not null;column:line_order"`
	LineType   LineType         `gorm:"type:varchar(50);not null;column:line_type"`
	SpeakerID  *uuid.UUID       `gorm:"type:uuid;column:speaker_id"`
	Text       *string          `gorm:"type:text"`
	Emotion    *string          `gorm:"type:text"`
	DurationMs *int             `gorm:"column:duration_ms"`
	SfxID      *uuid.UUID       `gorm:"type:uuid;column:sfx_id"`
	Volume     *decimal.Decimal `gorm:"type:decimal(3,2);column:volume"`
	AudioID    *uuid.UUID       `gorm:"type:uuid;column:audio_id"`
	CreatedAt  time.Time        `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt  time.Time        `gorm:"not null;default:CURRENT_TIMESTAMP"`

	// リレーション
	Episode Episode      `gorm:"foreignKey:EpisodeID"`
	Speaker *Character   `gorm:"foreignKey:SpeakerID"`
	Sfx     *SoundEffect `gorm:"foreignKey:SfxID"`
	Audio   *Audio       `gorm:"foreignKey:AudioID"`
}

// テーブル名を指定
func (ScriptLine) TableName() string {
	return "script_lines"
}
