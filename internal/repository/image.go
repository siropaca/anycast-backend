package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/logger"
	"github.com/siropaca/anycast-backend/internal/model"
)

// 画像データへのアクセスインターフェース
type ImageRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.Image, error)
	Delete(ctx context.Context, id uuid.UUID) error
	FindOrphaned(ctx context.Context) ([]model.Image, error)
}

type imageRepository struct {
	db *gorm.DB
}

// ImageRepository の実装を返す
func NewImageRepository(db *gorm.DB) ImageRepository {
	return &imageRepository{db: db}
}

// 指定された ID の画像を取得する
func (r *imageRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Image, error) {
	var image model.Image

	if err := r.db.WithContext(ctx).First(&image, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound.WithMessage("Image not found")
		}

		logger.FromContext(ctx).Error("failed to fetch image", "error", err, "image_id", id)
		return nil, apperror.ErrInternal.WithMessage("Failed to fetch image").WithError(err)
	}

	return &image, nil
}

// 画像を削除する
func (r *imageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.Image{}, "id = ?", id)
	if result.Error != nil {
		logger.FromContext(ctx).Error("failed to delete image", "error", result.Error, "id", id)
		return apperror.ErrInternal.WithMessage("Failed to delete image").WithError(result.Error)
	}

	if result.RowsAffected == 0 {
		return apperror.ErrNotFound.WithMessage("Image not found")
	}

	return nil
}

// どのテーブルからも参照されていない孤児レコードを取得する
// 対象: users.avatar_id, channels.artwork_id, episodes.artwork_id
// 条件: created_at から 1 時間以上経過したレコードのみ
func (r *imageRepository) FindOrphaned(ctx context.Context) ([]model.Image, error) {
	var images []model.Image

	query := `
		SELECT i.* FROM images i
		WHERE i.created_at < NOW() - INTERVAL '1 hour'
		AND NOT EXISTS (SELECT 1 FROM users u WHERE u.avatar_id = i.id)
		AND NOT EXISTS (SELECT 1 FROM channels c WHERE c.artwork_id = i.id)
		AND NOT EXISTS (SELECT 1 FROM episodes e WHERE e.artwork_id = i.id)
		ORDER BY i.created_at DESC
	`

	if err := r.db.WithContext(ctx).Raw(query).Scan(&images).Error; err != nil {
		logger.FromContext(ctx).Error("failed to fetch orphaned images", "error", err)
		return nil, apperror.ErrInternal.WithMessage("Failed to fetch orphaned images").WithError(err)
	}

	return images, nil
}
