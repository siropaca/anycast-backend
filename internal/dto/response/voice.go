package response

import "github.com/google/uuid"

// ボイスのレスポンス
type VoiceResponse struct {
	ID              uuid.UUID `json:"id"`
	Provider        string    `json:"provider"`
	ProviderVoiceID string    `json:"providerVoiceId"`
	Name            string    `json:"name"`
	Gender          string    `json:"gender"`
	IsActive        bool      `json:"isActive"`
}

// ボイス一覧のレスポンス
type VoiceListResponse struct {
	Data []VoiceResponse `json:"data"`
}

// ボイス単体のレスポンス
type VoiceDataResponse struct {
	Data VoiceResponse `json:"data"`
}
