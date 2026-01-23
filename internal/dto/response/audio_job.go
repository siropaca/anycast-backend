package response

import (
	"time"

	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// 音声生成ジョブのレスポンス
type AudioJobResponse struct {
	ID             uuid.UUID                `json:"id" validate:"required"`
	EpisodeID      uuid.UUID                `json:"episodeId" validate:"required"`
	Status         string                   `json:"status" validate:"required"`
	Progress       int                      `json:"progress" validate:"required"`
	VoiceStyle     string                   `json:"voiceStyle"`
	BgmVolumeDB    float64                  `json:"bgmVolumeDb"`
	FadeOutMs      int                      `json:"fadeOutMs"`
	PaddingStartMs int                      `json:"paddingStartMs"`
	PaddingEndMs   int                      `json:"paddingEndMs"`
	Episode        *AudioJobEpisodeResponse `json:"episode" extensions:"x-nullable"`
	ResultAudio    *AudioResponse           `json:"resultAudio" extensions:"x-nullable"`
	ErrorMessage   *string                  `json:"errorMessage" extensions:"x-nullable"`
	ErrorCode      *string                  `json:"errorCode" extensions:"x-nullable"`
	StartedAt      *time.Time               `json:"startedAt" extensions:"x-nullable"`
	CompletedAt    *time.Time               `json:"completedAt" extensions:"x-nullable"`
	CreatedAt      time.Time                `json:"createdAt" validate:"required"`
	UpdatedAt      time.Time                `json:"updatedAt" validate:"required"`
}

// 音声生成ジョブに含まれるエピソード情報
type AudioJobEpisodeResponse struct {
	ID      uuid.UUID                `json:"id" validate:"required"`
	Title   string                   `json:"title" validate:"required"`
	Channel *AudioJobChannelResponse `json:"channel" extensions:"x-nullable"`
}

// 音声生成ジョブに含まれるチャンネル情報
type AudioJobChannelResponse struct {
	ID   uuid.UUID `json:"id" validate:"required"`
	Name string    `json:"name" validate:"required"`
}

// 音声生成ジョブ一覧のレスポンス
type AudioJobListResponse struct {
	Data []AudioJobResponse `json:"data" validate:"required"`
}

// 音声生成ジョブ詳細のレスポンス
type AudioJobDataResponse struct {
	Data AudioJobResponse `json:"data" validate:"required"`
}
