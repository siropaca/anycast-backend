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

// ImageRepository は画像データへのアクセスインターフェース
type ImageRepository interface {
	Create(ctx context.Context, image *model.Image) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Image, error)
	Delete(ctx context.Context, id uuid.UUID) error
	FindOrphaned(ctx context.Context) ([]model.Image, error)
}

type imageRepository struct {
	db *gorm.DB
}

// NewImageRepository は ImageRepository の実装を返す
func NewImageRepository(db *gorm.DB) ImageRepository {
	return &imageRepository{db: db}
}

// Create は画像を作成する
func (r *imageRepository) Create(ctx context.Context, image *model.Image) error {
	if err := r.db.WithContext(ctx).Create(image).Error; err != nil {
		logger.FromContext(ctx).Error("failed to create image", "error", err)
		return apperror.ErrInternal.WithMessage("画像の作成に失敗しました").WithError(err)
	}

	return nil
}

// FindByID は指定された ID の画像を取得する
func (r *imageRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Image, error) {
	var image model.Image

	if err := r.db.WithContext(ctx).First(&image, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound.WithMessage("画像が見つかりません")
		}

		logger.FromContext(ctx).Error("failed to fetch image", "error", err, "image_id", id)
		return nil, apperror.ErrInternal.WithMessage("画像の取得に失敗しました").WithError(err)
	}

	return &image, nil
}

// Delete は画像を削除する
func (r *imageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.Image{}, "id = ?", id)
	if result.Error != nil {
		logger.FromContext(ctx).Error("failed to delete image", "error", result.Error, "id", id)
		return apperror.ErrInternal.WithMessage("画像の削除に失敗しました").WithError(result.Error)
	}

	if result.RowsAffected == 0 {
		return apperror.ErrNotFound.WithMessage("画像が見つかりません")
	}

	return nil
}

// FindOrphaned はどのテーブルからも参照されていない孤児レコードを取得する
//
// 対象: users.avatar_id, channels.artwork_id, episodes.artwork_id, characters.avatar_id
// 条件: created_at から 1 時間以上経過したレコードのみ
func (r *imageRepository) FindOrphaned(ctx context.Context) ([]model.Image, error) {
	var images []model.Image

	query := `
		SELECT i.* FROM images i
		WHERE i.created_at < NOW() - INTERVAL '1 hour'
		AND NOT EXISTS (SELECT 1 FROM users u WHERE u.avatar_id = i.id)
		AND NOT EXISTS (SELECT 1 FROM channels c WHERE c.artwork_id = i.id)
		AND NOT EXISTS (SELECT 1 FROM episodes e WHERE e.artwork_id = i.id)
		AND NOT EXISTS (SELECT 1 FROM characters ch WHERE ch.avatar_id = i.id)
		ORDER BY i.created_at DESC
	`

	if err := r.db.WithContext(ctx).Raw(query).Scan(&images).Error; err != nil {
		logger.FromContext(ctx).Error("failed to fetch orphaned images", "error", err)
		return nil, apperror.ErrInternal.WithMessage("孤立した画像の取得に失敗しました").WithError(err)
	}

	return images, nil
}
