package response

import (
	"time"

	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// 台本生成ジョブのレスポンス
type ScriptJobResponse struct {
	ID               uuid.UUID                 `json:"id" validate:"required"`
	EpisodeID        uuid.UUID                 `json:"episodeId" validate:"required"`
	Status           string                    `json:"status" validate:"required"`
	Progress         int                       `json:"progress" validate:"required"`
	Prompt           string                    `json:"prompt"`
	DurationMinutes  int                       `json:"durationMinutes"`
	WithEmotion      bool                      `json:"withEmotion"`
	Episode          *ScriptJobEpisodeResponse `json:"episode" extensions:"x-nullable"`
	ScriptLinesCount *int                      `json:"scriptLinesCount" extensions:"x-nullable"`
	ErrorMessage     *string                   `json:"errorMessage" extensions:"x-nullable"`
	ErrorCode        *string                   `json:"errorCode" extensions:"x-nullable"`
	StartedAt        *time.Time                `json:"startedAt" extensions:"x-nullable"`
	CompletedAt      *time.Time                `json:"completedAt" extensions:"x-nullable"`
	CreatedAt        time.Time                 `json:"createdAt" validate:"required"`
	UpdatedAt        time.Time                 `json:"updatedAt" validate:"required"`
}

// 台本生成ジョブに含まれるエピソード情報
type ScriptJobEpisodeResponse struct {
	ID      uuid.UUID                 `json:"id" validate:"required"`
	Title   string                    `json:"title" validate:"required"`
	Channel *ScriptJobChannelResponse `json:"channel" extensions:"x-nullable"`
}

// 台本生成ジョブに含まれるチャンネル情報
type ScriptJobChannelResponse struct {
	ID   uuid.UUID `json:"id" validate:"required"`
	Name string    `json:"name" validate:"required"`
}

// 台本生成ジョブ一覧のレスポンス
type ScriptJobListResponse struct {
	Data []ScriptJobResponse `json:"data" validate:"required"`
}

// 台本生成ジョブ詳細のレスポンス
type ScriptJobDataResponse struct {
	Data ScriptJobResponse `json:"data" validate:"required"`
}

// 台本生成ジョブ詳細のレスポンス（data が null の場合あり）
type ScriptJobDataNullableResponse struct {
	Data *ScriptJobResponse `json:"data" extensions:"x-nullable"`
}
