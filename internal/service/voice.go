package service

import (
	"context"
	"log/slog"

	"github.com/siropaca/anycast-backend/internal/logger"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/repository"
)

type voiceService struct {
	voiceRepo repository.VoiceRepository
}

// NewVoiceService は VoiceService の実装を返す
func NewVoiceService(voiceRepo repository.VoiceRepository) VoiceService {
	return &voiceService{voiceRepo: voiceRepo}
}

// ListVoices はフィルタ条件に基づいてボイス一覧を取得する
func (s *voiceService) ListVoices(ctx context.Context, filter repository.VoiceFilter) ([]model.Voice, error) {
	log := logger.FromContext(ctx)
	log.Debug("listing voices", slog.Any("filter", filter))

	voices, err := s.voiceRepo.FindAll(ctx, filter)
	if err != nil {
		log.Error("failed to list voices", slog.Any("error", err))
		return nil, err
	}

	log.Info("voices listed", slog.Int("count", len(voices)))
	return voices, nil
}

// GetVoice は指定された ID のボイスを取得する
func (s *voiceService) GetVoice(ctx context.Context, id string) (*model.Voice, error) {
	log := logger.FromContext(ctx)
	log.Debug("getting voice", slog.String("id", id))

	voice, err := s.voiceRepo.FindByID(ctx, id)
	if err != nil {
		log.Error("failed to get voice", slog.String("id", id), slog.Any("error", err))
		return nil, err
	}

	return voice, nil
}
