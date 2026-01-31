package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// ReactionRepository のモック
type mockReactionRepository struct {
	mock.Mock
}

func (m *mockReactionRepository) FindLikesByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]model.Reaction, int64, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]model.Reaction), args.Get(1).(int64), args.Error(2)
}

func TestNewReactionService(t *testing.T) {
	t.Run("ReactionService を作成できる", func(t *testing.T) {
		mockRepo := new(mockReactionRepository)
		mockStorage := new(mockStorageClient)
		svc := NewReactionService(mockRepo, mockStorage)

		assert.NotNil(t, svc)
	})
}

func TestReactionService_ListLikes(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("高評価したエピソード一覧を取得できる", func(t *testing.T) {
		mockRepo := new(mockReactionRepository)
		mockStorage := new(mockStorageClient)

		channelID := uuid.New()
		episodeID := uuid.New()
		now := time.Now()
		reactions := []model.Reaction{
			{
				ID:           uuid.New(),
				UserID:       userID,
				EpisodeID:    episodeID,
				ReactionType: model.ReactionTypeLike,
				CreatedAt:    now,
				Episode: model.Episode{
					ID:          episodeID,
					ChannelID:   channelID,
					Title:       "テストエピソード",
					Description: "テスト説明",
					PublishedAt: &now,
					Channel: model.Channel{
						ID:   channelID,
						Name: "テストチャンネル",
					},
				},
			},
		}
		mockRepo.On("FindLikesByUserID", mock.Anything, userID, 20, 0).Return(reactions, int64(1), nil)

		svc := NewReactionService(mockRepo, mockStorage)
		result, err := svc.ListLikes(ctx, userID.String(), 20, 0)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Data, 1)
		assert.Equal(t, episodeID, result.Data[0].Episode.ID)
		assert.Equal(t, "テストエピソード", result.Data[0].Episode.Title)
		assert.Equal(t, "テスト説明", result.Data[0].Episode.Description)
		assert.Equal(t, channelID, result.Data[0].Episode.Channel.ID)
		assert.Equal(t, "テストチャンネル", result.Data[0].Episode.Channel.Name)
		assert.Nil(t, result.Data[0].Episode.Channel.Artwork)
		assert.Equal(t, now, result.Data[0].LikedAt)
		assert.Equal(t, int64(1), result.Pagination.Total)
		assert.Equal(t, 20, result.Pagination.Limit)
		assert.Equal(t, 0, result.Pagination.Offset)
		mockRepo.AssertExpectations(t)
	})

	t.Run("アートワーク付きのチャンネルを取得できる", func(t *testing.T) {
		mockRepo := new(mockReactionRepository)
		mockStorage := new(mockStorageClient)

		artworkID := uuid.New()
		channelID := uuid.New()
		reactions := []model.Reaction{
			{
				ID:           uuid.New(),
				UserID:       userID,
				EpisodeID:    uuid.New(),
				ReactionType: model.ReactionTypeLike,
				CreatedAt:    time.Now(),
				Episode: model.Episode{
					ID:        uuid.New(),
					ChannelID: channelID,
					Title:     "エピソード",
					Channel: model.Channel{
						ID:        channelID,
						Name:      "チャンネル",
						ArtworkID: &artworkID,
						Artwork: &model.Image{
							ID:   artworkID,
							Path: "https://example.com/artwork.png",
						},
					},
				},
			},
		}
		mockRepo.On("FindLikesByUserID", mock.Anything, userID, 20, 0).Return(reactions, int64(1), nil)

		svc := NewReactionService(mockRepo, mockStorage)
		result, err := svc.ListLikes(ctx, userID.String(), 20, 0)

		assert.NoError(t, err)
		assert.NotNil(t, result.Data[0].Episode.Channel.Artwork)
		assert.Equal(t, artworkID, result.Data[0].Episode.Channel.Artwork.ID)
		assert.Equal(t, "https://example.com/artwork.png", result.Data[0].Episode.Channel.Artwork.URL)
		mockRepo.AssertExpectations(t)
	})

	t.Run("GCS パスのアートワークは署名付き URL が生成される", func(t *testing.T) {
		mockRepo := new(mockReactionRepository)
		mockStorage := new(mockStorageClient)

		artworkID := uuid.New()
		reactions := []model.Reaction{
			{
				ID:           uuid.New(),
				UserID:       userID,
				EpisodeID:    uuid.New(),
				ReactionType: model.ReactionTypeLike,
				CreatedAt:    time.Now(),
				Episode: model.Episode{
					ID:    uuid.New(),
					Title: "エピソード",
					Channel: model.Channel{
						ID:        uuid.New(),
						Name:      "チャンネル",
						ArtworkID: &artworkID,
						Artwork: &model.Image{
							ID:   artworkID,
							Path: "images/artwork.png",
						},
					},
				},
			},
		}
		mockRepo.On("FindLikesByUserID", mock.Anything, userID, 20, 0).Return(reactions, int64(1), nil)
		mockStorage.On("GenerateSignedURL", mock.Anything, "images/artwork.png", mock.Anything).Return("https://storage.example.com/signed-url", nil)

		svc := NewReactionService(mockRepo, mockStorage)
		result, err := svc.ListLikes(ctx, userID.String(), 20, 0)

		assert.NoError(t, err)
		assert.NotNil(t, result.Data[0].Episode.Channel.Artwork)
		assert.Equal(t, "https://storage.example.com/signed-url", result.Data[0].Episode.Channel.Artwork.URL)
		mockRepo.AssertExpectations(t)
		mockStorage.AssertExpectations(t)
	})

	t.Run("空の高評価一覧を取得できる", func(t *testing.T) {
		mockRepo := new(mockReactionRepository)
		mockStorage := new(mockStorageClient)

		mockRepo.On("FindLikesByUserID", mock.Anything, userID, 20, 0).Return([]model.Reaction{}, int64(0), nil)

		svc := NewReactionService(mockRepo, mockStorage)
		result, err := svc.ListLikes(ctx, userID.String(), 20, 0)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result.Data)
		assert.Equal(t, int64(0), result.Pagination.Total)
		mockRepo.AssertExpectations(t)
	})

	t.Run("無効な UUID の場合はエラーを返す", func(t *testing.T) {
		mockRepo := new(mockReactionRepository)
		mockStorage := new(mockStorageClient)

		svc := NewReactionService(mockRepo, mockStorage)
		result, err := svc.ListLikes(ctx, "invalid-uuid", 20, 0)

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("リポジトリがエラーを返すとエラーを返す", func(t *testing.T) {
		mockRepo := new(mockReactionRepository)
		mockStorage := new(mockStorageClient)

		mockRepo.On("FindLikesByUserID", mock.Anything, userID, 20, 0).Return(nil, int64(0), apperror.ErrInternal.WithMessage("Database error"))

		svc := NewReactionService(mockRepo, mockStorage)
		result, err := svc.ListLikes(ctx, userID.String(), 20, 0)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}
