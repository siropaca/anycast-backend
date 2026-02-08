package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/infrastructure/storage"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// CategoryRepository のモック
type mockCategoryRepository struct {
	mock.Mock
}

func (m *mockCategoryRepository) FindAllActive(ctx context.Context) ([]model.Category, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Category), args.Error(1)
}

func (m *mockCategoryRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Category, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Category), args.Error(1)
}

func (m *mockCategoryRepository) FindBySlug(ctx context.Context, slug string) (*model.Category, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Category), args.Error(1)
}

func (m *mockCategoryRepository) FindStatsByCategoryIDs(ctx context.Context, categoryIDs []uuid.UUID) (map[uuid.UUID]repository.CategoryStats, error) {
	args := m.Called(ctx, categoryIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[uuid.UUID]repository.CategoryStats), args.Error(1)
}

func TestNewCategoryService(t *testing.T) {
	t.Run("CategoryService を作成できる", func(t *testing.T) {
		mockRepo := new(mockCategoryRepository)
		mockStorage := new(mockStorageClient)
		svc := NewCategoryService(mockRepo, mockStorage)

		assert.NotNil(t, svc)
	})
}

func TestCategoryService_ListCategories(t *testing.T) {
	t.Run("カテゴリ一覧を取得できる", func(t *testing.T) {
		mockRepo := new(mockCategoryRepository)
		mockStorage := new(mockStorageClient)
		categories := []model.Category{
			{ID: uuid.New(), Slug: "technology", Name: "テクノロジー", SortOrder: 0, IsActive: true},
			{ID: uuid.New(), Slug: "news", Name: "ニュース", SortOrder: 1, IsActive: true},
		}
		statsMap := map[uuid.UUID]repository.CategoryStats{
			categories[0].ID: {CategoryID: categories[0].ID, ChannelCount: 5, EpisodeCount: 30},
			categories[1].ID: {CategoryID: categories[1].ID, ChannelCount: 3, EpisodeCount: 15},
		}
		mockRepo.On("FindAllActive", mock.Anything).Return(categories, nil)
		mockRepo.On("FindStatsByCategoryIDs", mock.Anything, []uuid.UUID{categories[0].ID, categories[1].ID}).Return(statsMap, nil)

		svc := NewCategoryService(mockRepo, mockStorage)
		result, err := svc.ListCategories(context.Background())

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "technology", result[0].Slug)
		assert.Equal(t, 5, result[0].ChannelCount)
		assert.Equal(t, 30, result[0].EpisodeCount)
		assert.Equal(t, "news", result[1].Slug)
		assert.Equal(t, 3, result[1].ChannelCount)
		assert.Equal(t, 15, result[1].EpisodeCount)
		mockRepo.AssertExpectations(t)
	})

	t.Run("画像付きカテゴリ一覧を取得できる", func(t *testing.T) {
		mockRepo := new(mockCategoryRepository)
		mockStorage := new(mockStorageClient)
		imageID := uuid.New()
		categories := []model.Category{
			{
				ID:      uuid.New(),
				Slug:    "technology",
				Name:    "テクノロジー",
				ImageID: &imageID,
				Image: &model.Image{
					ID:   imageID,
					Path: "images/tech.png",
				},
				SortOrder: 0,
				IsActive:  true,
			},
		}
		statsMap := map[uuid.UUID]repository.CategoryStats{
			categories[0].ID: {CategoryID: categories[0].ID, ChannelCount: 2, EpisodeCount: 10},
		}
		mockRepo.On("FindAllActive", mock.Anything).Return(categories, nil)
		mockRepo.On("FindStatsByCategoryIDs", mock.Anything, []uuid.UUID{categories[0].ID}).Return(statsMap, nil)
		mockStorage.On("GenerateSignedURL", mock.Anything, "images/tech.png", storage.SignedURLExpirationImage).Return("https://storage.example.com/signed/tech.png", nil)

		svc := NewCategoryService(mockRepo, mockStorage)
		result, err := svc.ListCategories(context.Background())

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.NotNil(t, result[0].Image)
		assert.Equal(t, imageID, result[0].Image.ID)
		assert.Equal(t, "https://storage.example.com/signed/tech.png", result[0].Image.URL)
		assert.Equal(t, 2, result[0].ChannelCount)
		assert.Equal(t, 10, result[0].EpisodeCount)
		mockRepo.AssertExpectations(t)
		mockStorage.AssertExpectations(t)
	})

	t.Run("外部 URL の画像は署名付き URL 生成をスキップする", func(t *testing.T) {
		mockRepo := new(mockCategoryRepository)
		mockStorage := new(mockStorageClient)
		imageID := uuid.New()
		categories := []model.Category{
			{
				ID:      uuid.New(),
				Slug:    "news",
				Name:    "ニュース",
				ImageID: &imageID,
				Image: &model.Image{
					ID:   imageID,
					Path: "https://example.com/news.png",
				},
				SortOrder: 0,
				IsActive:  true,
			},
		}
		statsMap := map[uuid.UUID]repository.CategoryStats{
			categories[0].ID: {CategoryID: categories[0].ID, ChannelCount: 1, EpisodeCount: 5},
		}
		mockRepo.On("FindAllActive", mock.Anything).Return(categories, nil)
		mockRepo.On("FindStatsByCategoryIDs", mock.Anything, []uuid.UUID{categories[0].ID}).Return(statsMap, nil)

		svc := NewCategoryService(mockRepo, mockStorage)
		result, err := svc.ListCategories(context.Background())

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.NotNil(t, result[0].Image)
		assert.Equal(t, "https://example.com/news.png", result[0].Image.URL)
		mockRepo.AssertExpectations(t)
		mockStorage.AssertNotCalled(t, "GenerateSignedURL")
	})

	t.Run("空のカテゴリ一覧を取得できる", func(t *testing.T) {
		mockRepo := new(mockCategoryRepository)
		mockStorage := new(mockStorageClient)
		mockRepo.On("FindAllActive", mock.Anything).Return([]model.Category{}, nil)
		mockRepo.On("FindStatsByCategoryIDs", mock.Anything, []uuid.UUID{}).Return(map[uuid.UUID]repository.CategoryStats{}, nil)

		svc := NewCategoryService(mockRepo, mockStorage)
		result, err := svc.ListCategories(context.Background())

		assert.NoError(t, err)
		assert.Empty(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("リポジトリがエラーを返すとエラーを返す", func(t *testing.T) {
		mockRepo := new(mockCategoryRepository)
		mockStorage := new(mockStorageClient)
		mockRepo.On("FindAllActive", mock.Anything).Return(nil, apperror.ErrInternal.WithMessage("Database error"))

		svc := NewCategoryService(mockRepo, mockStorage)
		result, err := svc.ListCategories(context.Background())

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("FindStatsByCategoryIDs がエラーを返すとエラーを返す", func(t *testing.T) {
		mockRepo := new(mockCategoryRepository)
		mockStorage := new(mockStorageClient)
		categories := []model.Category{
			{ID: uuid.New(), Slug: "technology", Name: "テクノロジー", SortOrder: 0, IsActive: true},
		}
		mockRepo.On("FindAllActive", mock.Anything).Return(categories, nil)
		mockRepo.On("FindStatsByCategoryIDs", mock.Anything, []uuid.UUID{categories[0].ID}).Return(nil, apperror.ErrInternal.WithMessage("Stats error"))

		svc := NewCategoryService(mockRepo, mockStorage)
		result, err := svc.ListCategories(context.Background())

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestCategoryService_GetCategoryBySlug(t *testing.T) {
	t.Run("スラッグ指定でカテゴリを取得できる", func(t *testing.T) {
		mockRepo := new(mockCategoryRepository)
		mockStorage := new(mockStorageClient)
		category := &model.Category{
			ID:        uuid.New(),
			Slug:      "technology",
			Name:      "テクノロジー",
			SortOrder: 0,
			IsActive:  true,
		}
		statsMap := map[uuid.UUID]repository.CategoryStats{
			category.ID: {CategoryID: category.ID, ChannelCount: 5, EpisodeCount: 30},
		}
		mockRepo.On("FindBySlug", mock.Anything, "technology").Return(category, nil)
		mockRepo.On("FindStatsByCategoryIDs", mock.Anything, []uuid.UUID{category.ID}).Return(statsMap, nil)

		svc := NewCategoryService(mockRepo, mockStorage)
		result, err := svc.GetCategoryBySlug(context.Background(), "technology")

		assert.NoError(t, err)
		assert.Equal(t, "technology", result.Slug)
		assert.Equal(t, "テクノロジー", result.Name)
		assert.Equal(t, 5, result.ChannelCount)
		assert.Equal(t, 30, result.EpisodeCount)
		mockRepo.AssertExpectations(t)
	})

	t.Run("画像付きカテゴリを取得できる", func(t *testing.T) {
		mockRepo := new(mockCategoryRepository)
		mockStorage := new(mockStorageClient)
		imageID := uuid.New()
		category := &model.Category{
			ID:      uuid.New(),
			Slug:    "technology",
			Name:    "テクノロジー",
			ImageID: &imageID,
			Image: &model.Image{
				ID:   imageID,
				Path: "images/tech.png",
			},
			SortOrder: 0,
			IsActive:  true,
		}
		statsMap := map[uuid.UUID]repository.CategoryStats{
			category.ID: {CategoryID: category.ID, ChannelCount: 4, EpisodeCount: 20},
		}
		mockRepo.On("FindBySlug", mock.Anything, "technology").Return(category, nil)
		mockRepo.On("FindStatsByCategoryIDs", mock.Anything, []uuid.UUID{category.ID}).Return(statsMap, nil)
		mockStorage.On("GenerateSignedURL", mock.Anything, "images/tech.png", storage.SignedURLExpirationImage).Return("https://storage.example.com/signed/tech.png", nil)

		svc := NewCategoryService(mockRepo, mockStorage)
		result, err := svc.GetCategoryBySlug(context.Background(), "technology")

		assert.NoError(t, err)
		assert.NotNil(t, result.Image)
		assert.Equal(t, imageID, result.Image.ID)
		assert.Equal(t, "https://storage.example.com/signed/tech.png", result.Image.URL)
		assert.Equal(t, 4, result.ChannelCount)
		assert.Equal(t, 20, result.EpisodeCount)
		mockRepo.AssertExpectations(t)
		mockStorage.AssertExpectations(t)
	})

	t.Run("リポジトリがエラーを返すとエラーを返す", func(t *testing.T) {
		mockRepo := new(mockCategoryRepository)
		mockStorage := new(mockStorageClient)
		mockRepo.On("FindBySlug", mock.Anything, "nonexistent").Return(nil, apperror.ErrNotFound.WithMessage("カテゴリが見つかりません"))

		svc := NewCategoryService(mockRepo, mockStorage)
		result, err := svc.GetCategoryBySlug(context.Background(), "nonexistent")

		assert.Error(t, err)
		assert.Empty(t, result.Slug)
		mockRepo.AssertExpectations(t)
	})

	t.Run("FindStatsByCategoryIDs がエラーを返すとエラーを返す", func(t *testing.T) {
		mockRepo := new(mockCategoryRepository)
		mockStorage := new(mockStorageClient)
		category := &model.Category{
			ID:        uuid.New(),
			Slug:      "technology",
			Name:      "テクノロジー",
			SortOrder: 0,
			IsActive:  true,
		}
		mockRepo.On("FindBySlug", mock.Anything, "technology").Return(category, nil)
		mockRepo.On("FindStatsByCategoryIDs", mock.Anything, []uuid.UUID{category.ID}).Return(nil, apperror.ErrInternal.WithMessage("Stats error"))

		svc := NewCategoryService(mockRepo, mockStorage)
		result, err := svc.GetCategoryBySlug(context.Background(), "technology")

		assert.Error(t, err)
		assert.Empty(t, result.Slug)
		mockRepo.AssertExpectations(t)
	})
}
