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
		mockRepo.On("FindAllActive", mock.Anything).Return(categories, nil)

		svc := NewCategoryService(mockRepo, mockStorage)
		result, err := svc.ListCategories(context.Background())

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "technology", result[0].Slug)
		assert.Equal(t, "news", result[1].Slug)
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
		mockRepo.On("FindAllActive", mock.Anything).Return(categories, nil)
		mockStorage.On("GenerateSignedURL", mock.Anything, "images/tech.png", storage.SignedURLExpirationImage).Return("https://storage.example.com/signed/tech.png", nil)

		svc := NewCategoryService(mockRepo, mockStorage)
		result, err := svc.ListCategories(context.Background())

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.NotNil(t, result[0].Image)
		assert.Equal(t, imageID, result[0].Image.ID)
		assert.Equal(t, "https://storage.example.com/signed/tech.png", result[0].Image.URL)
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
		mockRepo.On("FindAllActive", mock.Anything).Return(categories, nil)

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
}
