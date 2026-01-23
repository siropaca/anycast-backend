package model

import (
	"time"

	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// ScriptJobStatus は台本生成ジョブのステータスを表す
type ScriptJobStatus string

const (
	ScriptJobStatusPending    ScriptJobStatus = "pending"
	ScriptJobStatusProcessing ScriptJobStatus = "processing"
	ScriptJobStatusCompleted  ScriptJobStatus = "completed"
	ScriptJobStatusFailed     ScriptJobStatus = "failed"
)

// ScriptJob は非同期台本生成ジョブを表す
type ScriptJob struct {
	ID        uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EpisodeID uuid.UUID       `gorm:"type:uuid;not null;column:episode_id"`
	UserID    uuid.UUID       `gorm:"type:uuid;not null;column:user_id"`
	Status    ScriptJobStatus `gorm:"type:script_job_status;not null;default:'pending'"`
	Progress  int             `gorm:"not null;default:0"`

	// 生成パラメータ
	Prompt          string `gorm:"type:text;not null"`
	DurationMinutes int    `gorm:"not null;default:10;column:duration_minutes"`
	WithEmotion     bool   `gorm:"not null;default:false;column:with_emotion"`

	// 結果
	ErrorMessage *string `gorm:"type:text;column:error_message"`
	ErrorCode    *string `gorm:"type:varchar(50);column:error_code"`

	// タイムスタンプ
	StartedAt   *time.Time `gorm:"column:started_at"`
	CompletedAt *time.Time `gorm:"column:completed_at"`
	CreatedAt   time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP"`

	// リレーション
	Episode Episode `gorm:"foreignKey:EpisodeID"`
	User    User    `gorm:"foreignKey:UserID"`
}
