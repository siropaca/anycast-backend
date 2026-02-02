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

// CategoryRepository はカテゴリデータへのアクセスインターフェース
type CategoryRepository interface {
	FindAllActive(ctx context.Context) ([]model.Category, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Category, error)
}

type categoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository は CategoryRepository の実装を返す
func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

// FindAllActive はアクティブなカテゴリを表示順で取得する
func (r *categoryRepository) FindAllActive(ctx context.Context) ([]model.Category, error) {
	var categories []model.Category

	if err := r.db.WithContext(ctx).Preload("Image").Where("is_active = ?", true).Order("sort_order ASC").Find(&categories).Error; err != nil {
		logger.FromContext(ctx).Error("failed to fetch categories", "error", err)
		return nil, apperror.ErrInternal.WithMessage("カテゴリ一覧の取得に失敗しました").WithError(err)
	}

	return categories, nil
}

// FindByID は指定された ID のカテゴリを取得する
func (r *categoryRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Category, error) {
	var category model.Category

	if err := r.db.WithContext(ctx).Preload("Image").First(&category, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound.WithMessage("カテゴリが見つかりません")
		}
		logger.FromContext(ctx).Error("failed to fetch category", "error", err, "category_id", id)
		return nil, apperror.ErrInternal.WithMessage("カテゴリの取得に失敗しました").WithError(err)
	}

	return &category, nil
}
