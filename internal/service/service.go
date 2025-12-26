package service

import (
	"context"

	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// ボイス関連のビジネスロジックインターフェース
type VoiceService interface {
	ListVoices(ctx context.Context, filter repository.VoiceFilter) ([]model.Voice, error)
	GetVoice(ctx context.Context, id string) (*model.Voice, error)
}
