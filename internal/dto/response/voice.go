package response

import "github.com/google/uuid"

// VoiceResponse はボイスのレスポンス
type VoiceResponse struct {
	ID              uuid.UUID `json:"id"`
	Provider        string    `json:"provider"`
	ProviderVoiceID string    `json:"providerVoiceId"`
	Name            string    `json:"name"`
	Gender          *string   `json:"gender"`
	IsActive        bool      `json:"isActive"`
}
