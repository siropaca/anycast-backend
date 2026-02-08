package service

import (
	"context"

	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/infrastructure/storage"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// CategoryService はカテゴリ関連のビジネスロジックインターフェースを表す
type CategoryService interface {
	ListCategories(ctx context.Context) ([]response.CategoryResponse, error)
	GetCategoryBySlug(ctx context.Context, slug string) (response.CategoryResponse, error)
}

type categoryService struct {
	categoryRepo  repository.CategoryRepository
	storageClient storage.Client
}

// NewCategoryService は categoryService を生成して CategoryService として返す
func NewCategoryService(categoryRepo repository.CategoryRepository, storageClient storage.Client) CategoryService {
	return &categoryService{
		categoryRepo:  categoryRepo,
		storageClient: storageClient,
	}
}

// ListCategories はアクティブなカテゴリ一覧を取得する
func (s *categoryService) ListCategories(ctx context.Context) ([]response.CategoryResponse, error) {
	categories, err := s.categoryRepo.FindAllActive(ctx)
	if err != nil {
		return nil, err
	}

	categoryIDs := make([]uuid.UUID, len(categories))
	for i, c := range categories {
		categoryIDs[i] = c.ID
	}

	statsMap, err := s.categoryRepo.FindStatsByCategoryIDs(ctx, categoryIDs)
	if err != nil {
		return nil, err
	}

	result := make([]response.CategoryResponse, len(categories))
	for i, c := range categories {
		resp, err := s.toCategoryResponse(ctx, &c, statsMap[c.ID])
		if err != nil {
			return nil, err
		}
		result[i] = resp
	}

	return result, nil
}

// GetCategoryBySlug は指定されたスラッグのカテゴリを取得する
func (s *categoryService) GetCategoryBySlug(ctx context.Context, slug string) (response.CategoryResponse, error) {
	category, err := s.categoryRepo.FindBySlug(ctx, slug)
	if err != nil {
		return response.CategoryResponse{}, err
	}

	statsMap, err := s.categoryRepo.FindStatsByCategoryIDs(ctx, []uuid.UUID{category.ID})
	if err != nil {
		return response.CategoryResponse{}, err
	}

	return s.toCategoryResponse(ctx, category, statsMap[category.ID])
}

// toCategoryResponse は Category モデルをレスポンス DTO に変換する
func (s *categoryService) toCategoryResponse(ctx context.Context, c *model.Category, stats repository.CategoryStats) (response.CategoryResponse, error) {
	resp := response.CategoryResponse{
		ID:           c.ID,
		Slug:         c.Slug,
		Name:         c.Name,
		ChannelCount: stats.ChannelCount,
		EpisodeCount: stats.EpisodeCount,
		SortOrder:    c.SortOrder,
		IsActive:     c.IsActive,
	}

	if c.Image != nil {
		var imageURL string
		if storage.IsExternalURL(c.Image.Path) {
			imageURL = c.Image.Path
		} else {
			var err error
			imageURL, err = s.storageClient.GenerateSignedURL(ctx, c.Image.Path, storage.SignedURLExpirationImage)
			if err != nil {
				return response.CategoryResponse{}, err
			}
		}
		resp.Image = &response.ArtworkResponse{
			ID:  c.Image.ID,
			URL: imageURL,
		}
	}

	return resp, nil
}
