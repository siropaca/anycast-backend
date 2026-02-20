package model

import (
	"time"

	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// AudioJobStatus は音声生成ジョブのステータスを表す
type AudioJobStatus string

const (
	AudioJobStatusPending    AudioJobStatus = "pending"
	AudioJobStatusProcessing AudioJobStatus = "processing"
	AudioJobStatusCanceling  AudioJobStatus = "canceling"
	AudioJobStatusCompleted  AudioJobStatus = "completed"
	AudioJobStatusFailed     AudioJobStatus = "failed"
	AudioJobStatusCanceled   AudioJobStatus = "canceled"
)

// AudioJobType は音声生成ジョブの種別を表す
type AudioJobType string

const (
	AudioJobTypeVoice AudioJobType = "voice"
	AudioJobTypeFull  AudioJobType = "full"
	AudioJobTypeRemix AudioJobType = "remix"
)

// AudioJob は非同期音声生成ジョブを表す
type AudioJob struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EpisodeID uuid.UUID      `gorm:"type:uuid;not null;column:episode_id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;column:user_id"`
	Status    AudioJobStatus `gorm:"type:audio_job_status;not null;default:'pending'"`
	JobType   AudioJobType   `gorm:"type:audio_job_type;not null;default:'voice';column:job_type"`
	Progress  int            `gorm:"not null;default:0"`

	// BGM 参照
	BgmID       *uuid.UUID `gorm:"type:uuid;column:bgm_id"`
	SystemBgmID *uuid.UUID `gorm:"type:uuid;column:system_bgm_id"`

	// BGM ミキシング設定
	BgmVolumeDB    float64 `gorm:"type:decimal(5,2);not null;default:-20.0;column:bgm_volume_db"`
	FadeOutMs      int     `gorm:"not null;default:3000;column:fade_out_ms"`
	PaddingStartMs int     `gorm:"not null;default:1000;column:padding_start_ms"`
	PaddingEndMs   int     `gorm:"not null;default:3000;column:padding_end_ms"`

	// 結果
	ResultAudioID *uuid.UUID `gorm:"type:uuid;column:result_audio_id"`
	ErrorMessage  *string    `gorm:"type:text;column:error_message"`
	ErrorCode     *string    `gorm:"type:varchar(50);column:error_code"`

	// タイムスタンプ
	StartedAt   *time.Time `gorm:"column:started_at"`
	CompletedAt *time.Time `gorm:"column:completed_at"`
	CreatedAt   time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP"`

	// リレーション
	Episode     Episode    `gorm:"foreignKey:EpisodeID"`
	User        User       `gorm:"foreignKey:UserID"`
	Bgm         *Bgm       `gorm:"foreignKey:BgmID"`
	SystemBgm   *SystemBgm `gorm:"foreignKey:SystemBgmID"`
	ResultAudio *Audio     `gorm:"foreignKey:ResultAudioID"`
}
