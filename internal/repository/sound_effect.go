package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
)

// 効果音データへのアクセスインターフェース
type SoundEffectRepository interface {
	FindByName(ctx context.Context, name string) (*model.SoundEffect, error)
	FindAll(ctx context.Context) ([]model.SoundEffect, error)
}

type soundEffectRepository struct {
	db *gorm.DB
}

// SoundEffectRepository の実装を返す
func NewSoundEffectRepository(db *gorm.DB) SoundEffectRepository {
	return &soundEffectRepository{db: db}
}

// 名前で効果音を取得する
func (r *soundEffectRepository) FindByName(ctx context.Context, name string) (*model.SoundEffect, error) {
	var sfx model.SoundEffect

	if err := r.db.WithContext(ctx).
		Preload("Audio").
		First(&sfx, "name = ?", name).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperror.ErrNotFound.WithMessage("Sound effect not found: " + name)
		}
		logger.FromContext(ctx).Error("failed to fetch sound effect", "error", err, "name", name)
		return nil, apperror.ErrInternal.WithMessage("Failed to fetch sound effect").WithError(err)
	}

	return &sfx, nil
}

// 全ての効果音を取得する
func (r *soundEffectRepository) FindAll(ctx context.Context) ([]model.SoundEffect, error) {
	var sfxList []model.SoundEffect

	if err := r.db.WithContext(ctx).
		Preload("Audio").
		Order("name ASC").
		Find(&sfxList).Error; err != nil {
		logger.FromContext(ctx).Error("failed to fetch sound effects", "error", err)
		return nil, apperror.ErrInternal.WithMessage("Failed to fetch sound effects").WithError(err)
	}

	return sfxList, nil
}
