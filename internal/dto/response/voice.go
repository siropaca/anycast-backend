package response

import "github.com/siropaca/anycast-backend/internal/pkg/uuid"

// ボイスのレスポンス
type VoiceResponse struct {
	ID              uuid.UUID `json:"id" validate:"required"`
	Provider        string    `json:"provider" validate:"required"`
	ProviderVoiceID string    `json:"providerVoiceId" validate:"required"`
	Name            string    `json:"name" validate:"required"`
	Gender          string    `json:"gender" validate:"required"`
	SampleAudioUrl  string    `json:"sampleAudioUrl" validate:"required"`
	IsActive        bool      `json:"isActive" validate:"required"`
}

// ボイス一覧のレスポンス
type VoiceListResponse struct {
	Data []VoiceResponse `json:"data" validate:"required"`
}

// ボイス単体のレスポンス
type VoiceDataResponse struct {
	Data VoiceResponse `json:"data" validate:"required"`
}
