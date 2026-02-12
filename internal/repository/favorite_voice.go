package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// FavoriteVoiceRepository はボイスお気に入りのリポジトリインターフェースを表す
type FavoriteVoiceRepository interface {
	Create(ctx context.Context, fav *model.FavoriteVoice) error
	DeleteByUserIDAndVoiceID(ctx context.Context, userID, voiceID uuid.UUID) error
	FindVoiceIDsByUserID(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
	ExistsByUserIDAndVoiceID(ctx context.Context, userID, voiceID uuid.UUID) (bool, error)
}

type favoriteVoiceRepository struct {
	db *gorm.DB
}

// NewFavoriteVoiceRepository は FavoriteVoiceRepository の実装を返す
func NewFavoriteVoiceRepository(db *gorm.DB) FavoriteVoiceRepository {
	return &favoriteVoiceRepository{db: db}
}

// Create はボイスお気に入りを登録する
func (r *favoriteVoiceRepository) Create(ctx context.Context, fav *model.FavoriteVoice) error {
	if err := r.db.WithContext(ctx).Create(fav).Error; err != nil {
		logger.FromContext(ctx).Error("failed to create favorite voice", "error", err)
		return apperror.ErrInternal.WithMessage("お気に入りの登録に失敗しました").WithError(err)
	}
	return nil
}

// DeleteByUserIDAndVoiceID はユーザーとボイスに対応するお気に入りを削除する
func (r *favoriteVoiceRepository) DeleteByUserIDAndVoiceID(ctx context.Context, userID, voiceID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND voice_id = ?", userID, voiceID).
		Delete(&model.FavoriteVoice{})
	if result.Error != nil {
		logger.FromContext(ctx).Error("failed to delete favorite voice", "error", result.Error)
		return apperror.ErrInternal.WithMessage("お気に入りの解除に失敗しました").WithError(result.Error)
	}
	if result.RowsAffected == 0 {
		return apperror.ErrNotFound.WithMessage("お気に入りが見つかりません")
	}
	return nil
}

// FindVoiceIDsByUserID はユーザーがお気に入り登録しているボイス ID 一覧を取得する
func (r *favoriteVoiceRepository) FindVoiceIDsByUserID(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	var voiceIDs []uuid.UUID
	if err := r.db.WithContext(ctx).
		Model(&model.FavoriteVoice{}).
		Where("user_id = ?", userID).
		Pluck("voice_id", &voiceIDs).Error; err != nil {
		logger.FromContext(ctx).Error("failed to fetch favorite voice IDs", "error", err)
		return nil, apperror.ErrInternal.WithMessage("お気に入り一覧の取得に失敗しました").WithError(err)
	}
	return voiceIDs, nil
}

// ExistsByUserIDAndVoiceID はユーザーとボイスの組み合わせでお気に入りが存在するか確認する
func (r *favoriteVoiceRepository) ExistsByUserIDAndVoiceID(ctx context.Context, userID, voiceID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.FavoriteVoice{}).
		Where("user_id = ? AND voice_id = ?", userID, voiceID).
		Count(&count).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		logger.FromContext(ctx).Error("failed to check favorite voice existence", "error", err)
		return false, apperror.ErrInternal.WithMessage("お気に入りの確認に失敗しました").WithError(err)
	}
	return count > 0, nil
}
