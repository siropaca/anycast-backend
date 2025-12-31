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
