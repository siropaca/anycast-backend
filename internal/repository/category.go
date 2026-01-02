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

// カテゴリデータへのアクセスインターフェース
type CategoryRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.Category, error)
}

type categoryRepository struct {
	db *gorm.DB
}

// CategoryRepository の実装を返す
func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

// 指定された ID のカテゴリを取得する
func (r *categoryRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Category, error) {
	var category model.Category

	if err := r.db.WithContext(ctx).First(&category, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound.WithMessage("Category not found")
		}
		logger.FromContext(ctx).Error("failed to fetch category", "error", err, "category_id", id)
		return nil, apperror.ErrInternal.WithMessage("Failed to fetch category").WithError(err)
	}

	return &category, nil
}
