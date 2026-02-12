package service

import (
	"context"
	"slices"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// VoiceService はボイス関連のビジネスロジックインターフェースを表す
type VoiceService interface {
	ListVoices(ctx context.Context, userID string, filter repository.VoiceFilter) ([]model.Voice, []uuid.UUID, error)
	GetVoice(ctx context.Context, userID string, id string) (*model.Voice, bool, error)
	AddFavorite(ctx context.Context, userID string, voiceID string) (*model.FavoriteVoice, error)
	RemoveFavorite(ctx context.Context, userID string, voiceID string) error
}

type voiceService struct {
	voiceRepo    repository.VoiceRepository
	favVoiceRepo repository.FavoriteVoiceRepository
}

// NewVoiceService は voiceService を生成して VoiceService として返す
func NewVoiceService(voiceRepo repository.VoiceRepository, favVoiceRepo repository.FavoriteVoiceRepository) VoiceService {
	return &voiceService{voiceRepo: voiceRepo, favVoiceRepo: favVoiceRepo}
}

// ListVoices はフィルタ条件に基づいてボイス一覧を取得する（お気に入りボイス ID 一覧付き）
func (s *voiceService) ListVoices(ctx context.Context, userID string, filter repository.VoiceFilter) ([]model.Voice, []uuid.UUID, error) {
	voices, err := s.voiceRepo.FindAll(ctx, filter)
	if err != nil {
		return nil, nil, err
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, nil, err
	}

	favIDs, err := s.favVoiceRepo.FindVoiceIDsByUserID(ctx, uid)
	if err != nil {
		return nil, nil, err
	}

	// お気に入りを先頭にソート（安定ソート）
	slices.SortStableFunc(voices, func(a, b model.Voice) int {
		aFav := containsVoiceUUID(favIDs, a.ID)
		bFav := containsVoiceUUID(favIDs, b.ID)
		if aFav && !bFav {
			return -1
		}
		if !aFav && bFav {
			return 1
		}
		return 0
	})

	return voices, favIDs, nil
}

// GetVoice は指定された ID のボイスを取得する（お気に入り状態付き）
func (s *voiceService) GetVoice(ctx context.Context, userID string, id string) (*model.Voice, bool, error) {
	voice, err := s.voiceRepo.FindByID(ctx, id)
	if err != nil {
		return nil, false, err
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, false, err
	}

	vid, err := uuid.Parse(id)
	if err != nil {
		return nil, false, err
	}

	isFav, err := s.favVoiceRepo.ExistsByUserIDAndVoiceID(ctx, uid, vid)
	if err != nil {
		return nil, false, err
	}

	return voice, isFav, nil
}

// AddFavorite はボイスをお気に入りに登録する
func (s *voiceService) AddFavorite(ctx context.Context, userID string, voiceID string) (*model.FavoriteVoice, error) {
	// ボイスの存在チェック
	if _, err := s.voiceRepo.FindActiveByID(ctx, voiceID); err != nil {
		return nil, err
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	vid, err := uuid.Parse(voiceID)
	if err != nil {
		return nil, err
	}

	// 重複チェック
	exists, err := s.favVoiceRepo.ExistsByUserIDAndVoiceID(ctx, uid, vid)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperror.ErrAlreadyFavorited
	}

	fav := &model.FavoriteVoice{
		UserID:  uid,
		VoiceID: vid,
	}
	if err := s.favVoiceRepo.Create(ctx, fav); err != nil {
		return nil, err
	}

	return fav, nil
}

// RemoveFavorite はボイスのお気に入りを解除する
func (s *voiceService) RemoveFavorite(ctx context.Context, userID string, voiceID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	vid, err := uuid.Parse(voiceID)
	if err != nil {
		return err
	}

	return s.favVoiceRepo.DeleteByUserIDAndVoiceID(ctx, uid, vid)
}

// containsVoiceUUID はスライスに指定された UUID が含まれるかを返す
func containsVoiceUUID(ids []uuid.UUID, id uuid.UUID) bool {
	for _, v := range ids {
		if v == id {
			return true
		}
	}
	return false
}
