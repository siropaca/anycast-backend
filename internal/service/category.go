package service

import (
	"context"

	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// カテゴリ関連のビジネスロジックインターフェース
type CategoryService interface {
	ListCategories(ctx context.Context) ([]model.Category, error)
}

type categoryService struct {
	categoryRepo repository.CategoryRepository
}

// CategoryService の実装を返す
func NewCategoryService(categoryRepo repository.CategoryRepository) CategoryService {
	return &categoryService{categoryRepo: categoryRepo}
}

// アクティブなカテゴリ一覧を取得する
func (s *categoryService) ListCategories(ctx context.Context) ([]model.Category, error) {
	return s.categoryRepo.FindAllActive(ctx)
}
