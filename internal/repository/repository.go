package repository

import (
	"context"

	"github.com/siropaca/anycast-backend/internal/model"
)

// VoiceRepository はボイスデータへのアクセスインターフェース
type VoiceRepository interface {
	FindAll(ctx context.Context, filter VoiceFilter) ([]model.Voice, error)
	FindByID(ctx context.Context, id string) (*model.Voice, error)
}

// VoiceFilter はボイス検索のフィルタ条件
type VoiceFilter struct {
	Provider   *string
	Gender     *string
	ActiveOnly bool
}
