package repository

import (
	"context"

	"github.com/siropaca/anycast-backend/internal/model"
)

// ボイスデータへのアクセスインターフェース
type VoiceRepository interface {
	FindAll(ctx context.Context, filter VoiceFilter) ([]model.Voice, error)
	FindByID(ctx context.Context, id string) (*model.Voice, error)
}

// ボイス検索のフィルタ条件
type VoiceFilter struct {
	Provider *string
	Gender   *string
}
