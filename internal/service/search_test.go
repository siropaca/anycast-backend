package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

func TestSearchService_SearchChannels(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	categoryID := uuid.New()

	t.Run("キーワードでチャンネルを検索できる", func(t *testing.T) {
		mockChannelRepo := new(mockChannelRepository)
		mockStorage := new(mockStorageClient)

		channels := []model.Channel{
			{
				ID:          uuid.New(),
				Name:        "テックチャンネル",
				Description: "テクノロジーに関する話題",
				CategoryID:  categoryID,
				Category: model.Category{
					ID:       categoryID,
					Slug:     "technology",
					Name:     "テクノロジー",
					IsActive: true,
				},
				PublishedAt: &now,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			{
				ID:          uuid.New(),
				Name:        "テックニュース",
				Description: "最新のテック情報",
				CategoryID:  categoryID,
				Category: model.Category{
					ID:       categoryID,
					Slug:     "technology",
					Name:     "テクノロジー",
					IsActive: true,
				},
				PublishedAt: &now,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		}

		filter := repository.SearchChannelFilter{
			Query:  "テック",
			Limit:  20,
			Offset: 0,
		}
		mockChannelRepo.On("Search", mock.Anything, filter).Return(channels, int64(2), nil)

		svc := NewSearchService(mockChannelRepo, new(mockEpisodeRepository), mockStorage)
		result, err := svc.SearchChannels(ctx, filter)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Data, 2)
		assert.Equal(t, int64(2), result.Pagination.Total)
		assert.Equal(t, 20, result.Pagination.Limit)
		assert.Equal(t, 0, result.Pagination.Offset)
		assert.Equal(t, "テックチャンネル", result.Data[0].Name)
		assert.Equal(t, "テックニュース", result.Data[1].Name)
		mockChannelRepo.AssertExpectations(t)
	})

	t.Run("カテゴリスラッグでフィルタできる", func(t *testing.T) {
		mockChannelRepo := new(mockChannelRepository)
		mockStorage := new(mockStorageClient)

		channels := []model.Channel{
			{
				ID:          uuid.New(),
				Name:        "テックチャンネル",
				Description: "テクノロジーに関する話題",
				CategoryID:  categoryID,
				Category: model.Category{
					ID:       categoryID,
					Slug:     "technology",
					Name:     "テクノロジー",
					IsActive: true,
				},
				PublishedAt: &now,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		}

		slug := "technology"
		filter := repository.SearchChannelFilter{
			Query:        "テック",
			CategorySlug: &slug,
			Limit:        20,
			Offset:       0,
		}
		mockChannelRepo.On("Search", mock.Anything, filter).Return(channels, int64(1), nil)

		svc := NewSearchService(mockChannelRepo, new(mockEpisodeRepository), mockStorage)
		result, err := svc.SearchChannels(ctx, filter)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Data, 1)
		assert.Equal(t, int64(1), result.Pagination.Total)
		mockChannelRepo.AssertExpectations(t)
	})

	t.Run("アートワーク付きのチャンネルを検索できる", func(t *testing.T) {
		mockChannelRepo := new(mockChannelRepository)
		mockStorage := new(mockStorageClient)

		artworkID := uuid.New()
		channels := []model.Channel{
			{
				ID:          uuid.New(),
				Name:        "テックチャンネル",
				Description: "テクノロジーに関する話題",
				CategoryID:  categoryID,
				Category: model.Category{
					ID:       categoryID,
					Slug:     "technology",
					Name:     "テクノロジー",
					IsActive: true,
				},
				Artwork: &model.Image{
					ID:   artworkID,
					Path: "images/artwork.png",
				},
				PublishedAt: &now,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		}

		filter := repository.SearchChannelFilter{
			Query:  "テック",
			Limit:  20,
			Offset: 0,
		}
		mockChannelRepo.On("Search", mock.Anything, filter).Return(channels, int64(1), nil)
		mockStorage.On("GenerateSignedURL", mock.Anything, "images/artwork.png", mock.Anything).Return("https://signed-url.example.com/artwork.png", nil)

		svc := NewSearchService(mockChannelRepo, new(mockEpisodeRepository), mockStorage)
		result, err := svc.SearchChannels(ctx, filter)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Data, 1)
		assert.NotNil(t, result.Data[0].Artwork)
		assert.Equal(t, artworkID, result.Data[0].Artwork.ID)
		assert.Equal(t, "https://signed-url.example.com/artwork.png", result.Data[0].Artwork.URL)
		mockChannelRepo.AssertExpectations(t)
		mockStorage.AssertExpectations(t)
	})

	t.Run("外部 URL のアートワークは署名なしで返す", func(t *testing.T) {
		mockChannelRepo := new(mockChannelRepository)
		mockStorage := new(mockStorageClient)

		artworkID := uuid.New()
		channels := []model.Channel{
			{
				ID:          uuid.New(),
				Name:        "テックチャンネル",
				Description: "テクノロジーに関する話題",
				CategoryID:  categoryID,
				Category: model.Category{
					ID:       categoryID,
					Slug:     "technology",
					Name:     "テクノロジー",
					IsActive: true,
				},
				Artwork: &model.Image{
					ID:   artworkID,
					Path: "https://external.example.com/artwork.png",
				},
				PublishedAt: &now,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		}

		filter := repository.SearchChannelFilter{
			Query:  "テック",
			Limit:  20,
			Offset: 0,
		}
		mockChannelRepo.On("Search", mock.Anything, filter).Return(channels, int64(1), nil)

		svc := NewSearchService(mockChannelRepo, new(mockEpisodeRepository), mockStorage)
		result, err := svc.SearchChannels(ctx, filter)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Data[0].Artwork)
		assert.Equal(t, "https://external.example.com/artwork.png", result.Data[0].Artwork.URL)
		mockChannelRepo.AssertExpectations(t)
	})

	t.Run("検索結果が空の場合は空配列を返す", func(t *testing.T) {
		mockChannelRepo := new(mockChannelRepository)
		mockStorage := new(mockStorageClient)

		filter := repository.SearchChannelFilter{
			Query:  "存在しないキーワード",
			Limit:  20,
			Offset: 0,
		}
		mockChannelRepo.On("Search", mock.Anything, filter).Return([]model.Channel{}, int64(0), nil)

		svc := NewSearchService(mockChannelRepo, new(mockEpisodeRepository), mockStorage)
		result, err := svc.SearchChannels(ctx, filter)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Data, 0)
		assert.Equal(t, int64(0), result.Pagination.Total)
		mockChannelRepo.AssertExpectations(t)
	})

	t.Run("リポジトリエラー時はエラーを返す", func(t *testing.T) {
		mockChannelRepo := new(mockChannelRepository)
		mockStorage := new(mockStorageClient)

		filter := repository.SearchChannelFilter{
			Query:  "テック",
			Limit:  20,
			Offset: 0,
		}
		mockChannelRepo.On("Search", mock.Anything, filter).Return(nil, int64(0), errors.New("db error"))

		svc := NewSearchService(mockChannelRepo, new(mockEpisodeRepository), mockStorage)
		result, err := svc.SearchChannels(ctx, filter)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockChannelRepo.AssertExpectations(t)
	})
}

func TestSearchService_SearchEpisodes(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	channelID := uuid.New()

	t.Run("キーワードでエピソードを検索できる", func(t *testing.T) {
		mockEpisodeRepo := new(mockEpisodeRepository)
		mockStorage := new(mockStorageClient)

		episodes := []model.Episode{
			{
				ID:          uuid.New(),
				ChannelID:   channelID,
				Title:       "AI の最新動向",
				Description: "人工知能に関する最新ニュース",
				Channel: model.Channel{
					ID:   channelID,
					Name: "テックチャンネル",
				},
				PublishedAt: &now,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			{
				ID:          uuid.New(),
				ChannelID:   channelID,
				Title:       "AI と未来",
				Description: "AI がもたらす未来について",
				Channel: model.Channel{
					ID:   channelID,
					Name: "テックチャンネル",
				},
				PublishedAt: &now,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		}

		filter := repository.SearchEpisodeFilter{
			Query:  "AI",
			Limit:  20,
			Offset: 0,
		}
		mockEpisodeRepo.On("Search", mock.Anything, filter).Return(episodes, int64(2), nil)

		svc := NewSearchService(new(mockChannelRepository), mockEpisodeRepo, mockStorage)
		result, err := svc.SearchEpisodes(ctx, filter)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Data, 2)
		assert.Equal(t, int64(2), result.Pagination.Total)
		assert.Equal(t, 20, result.Pagination.Limit)
		assert.Equal(t, 0, result.Pagination.Offset)
		assert.Equal(t, "AI の最新動向", result.Data[0].Title)
		assert.Equal(t, channelID, result.Data[0].Channel.ID)
		assert.Equal(t, "テックチャンネル", result.Data[0].Channel.Name)
		mockEpisodeRepo.AssertExpectations(t)
	})

	t.Run("検索結果が空の場合は空配列を返す", func(t *testing.T) {
		mockEpisodeRepo := new(mockEpisodeRepository)
		mockStorage := new(mockStorageClient)

		filter := repository.SearchEpisodeFilter{
			Query:  "存在しないキーワード",
			Limit:  20,
			Offset: 0,
		}
		mockEpisodeRepo.On("Search", mock.Anything, filter).Return([]model.Episode{}, int64(0), nil)

		svc := NewSearchService(new(mockChannelRepository), mockEpisodeRepo, mockStorage)
		result, err := svc.SearchEpisodes(ctx, filter)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Data, 0)
		assert.Equal(t, int64(0), result.Pagination.Total)
		mockEpisodeRepo.AssertExpectations(t)
	})

	t.Run("リポジトリエラー時はエラーを返す", func(t *testing.T) {
		mockEpisodeRepo := new(mockEpisodeRepository)
		mockStorage := new(mockStorageClient)

		filter := repository.SearchEpisodeFilter{
			Query:  "AI",
			Limit:  20,
			Offset: 0,
		}
		mockEpisodeRepo.On("Search", mock.Anything, filter).Return(nil, int64(0), errors.New("db error"))

		svc := NewSearchService(new(mockChannelRepository), mockEpisodeRepo, mockStorage)
		result, err := svc.SearchEpisodes(ctx, filter)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockEpisodeRepo.AssertExpectations(t)
	})
}
