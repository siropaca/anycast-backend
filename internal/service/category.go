package service

import (
	"context"

	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// CategoryService はカテゴリ関連のビジネスロジックインターフェースを表す
type CategoryService interface {
	ListCategories(ctx context.Context) ([]model.Category, error)
}

type categoryService struct {
	categoryRepo repository.CategoryRepository
}

// NewCategoryService は categoryService を生成して CategoryService として返す
func NewCategoryService(categoryRepo repository.CategoryRepository) CategoryService {
	return &categoryService{categoryRepo: categoryRepo}
}

// ListCategories はアクティブなカテゴリ一覧を取得する
func (s *categoryService) ListCategories(ctx context.Context) ([]model.Category, error) {
	return s.categoryRepo.FindAllActive(ctx)
}
