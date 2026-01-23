package service

import (
	"context"

	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// VoiceService はボイス関連のビジネスロジックインターフェースを表す
type VoiceService interface {
	ListVoices(ctx context.Context, filter repository.VoiceFilter) ([]model.Voice, error)
	GetVoice(ctx context.Context, id string) (*model.Voice, error)
}

type voiceService struct {
	voiceRepo repository.VoiceRepository
}

// NewVoiceService は voiceService を生成して VoiceService として返す
func NewVoiceService(voiceRepo repository.VoiceRepository) VoiceService {
	return &voiceService{voiceRepo: voiceRepo}
}

// ListVoices はフィルタ条件に基づいてボイス一覧を取得する
func (s *voiceService) ListVoices(ctx context.Context, filter repository.VoiceFilter) ([]model.Voice, error) {
	return s.voiceRepo.FindAll(ctx, filter)
}

// GetVoice は指定された ID のボイスを取得する
func (s *voiceService) GetVoice(ctx context.Context, id string) (*model.Voice, error) {
	return s.voiceRepo.FindByID(ctx, id)
}
