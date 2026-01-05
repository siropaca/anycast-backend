package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/logger"
	"github.com/siropaca/anycast-backend/internal/model"
)

// 音声データへのアクセスインターフェース
type AudioRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.Audio, error)
	Create(ctx context.Context, audio *model.Audio) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type audioRepository struct {
	db *gorm.DB
}

// AudioRepository の実装を返す
func NewAudioRepository(db *gorm.DB) AudioRepository {
	return &audioRepository{db: db}
}

// ID で音声を取得する
func (r *audioRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Audio, error) {
	var audio model.Audio

	if err := r.db.WithContext(ctx).First(&audio, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.ErrNotFound.WithMessage("Audio not found")
		}

		logger.FromContext(ctx).Error("failed to fetch audio", "error", err, "id", id)
		return nil, apperror.ErrInternal.WithMessage("Failed to fetch audio").WithError(err)
	}

	return &audio, nil
}

// 音声を作成する
func (r *audioRepository) Create(ctx context.Context, audio *model.Audio) error {
	if err := r.db.WithContext(ctx).Create(audio).Error; err != nil {
		logger.FromContext(ctx).Error("failed to create audio", "error", err)
		return apperror.ErrInternal.WithMessage("Failed to create audio").WithError(err)
	}

	return nil
}

// 音声を削除する
func (r *audioRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.Audio{}, "id = ?", id)
	if result.Error != nil {
		logger.FromContext(ctx).Error("failed to delete audio", "error", result.Error, "id", id)
		return apperror.ErrInternal.WithMessage("Failed to delete audio").WithError(result.Error)
	}

	if result.RowsAffected == 0 {
		return apperror.ErrNotFound.WithMessage("Audio not found")
	}

	return nil
}
