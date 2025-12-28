package service

import (
	"context"

	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/repository"
)

type voiceService struct {
	voiceRepo repository.VoiceRepository
}

// VoiceService の実装を返す
func NewVoiceService(voiceRepo repository.VoiceRepository) VoiceService {
	return &voiceService{voiceRepo: voiceRepo}
}

// フィルタ条件に基づいてボイス一覧を取得する
func (s *voiceService) ListVoices(ctx context.Context, filter repository.VoiceFilter) ([]model.Voice, error) {
	return s.voiceRepo.FindAll(ctx, filter)
}

// 指定された ID のボイスを取得する
func (s *voiceService) GetVoice(ctx context.Context, id string) (*model.Voice, error) {
	return s.voiceRepo.FindByID(ctx, id)
}
