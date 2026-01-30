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

// FollowRepository のモック
type mockFollowRepository struct {
	mock.Mock
}

func (m *mockFollowRepository) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]model.Follow, int64, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]model.Follow), args.Get(1).(int64), args.Error(2)
}

func TestNewFollowService(t *testing.T) {
	t.Run("FollowService を作成できる", func(t *testing.T) {
		mockRepo := new(mockFollowRepository)
		mockStorage := new(mockStorageClient)
		svc := NewFollowService(mockRepo, mockStorage)

		assert.NotNil(t, svc)
	})
}

func TestFollowService_ListFollows(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("フォロー中のユーザー一覧を取得できる", func(t *testing.T) {
		mockRepo := new(mockFollowRepository)
		mockStorage := new(mockStorageClient)

		targetUser := model.User{
			ID:          uuid.New(),
			Username:    "target_user",
			DisplayName: "Target User",
		}
		follows := []model.Follow{
			{
				ID:           uuid.New(),
				UserID:       userID,
				TargetUserID: targetUser.ID,
				CreatedAt:    time.Now(),
				TargetUser:   targetUser,
			},
		}
		mockRepo.On("FindByUserID", mock.Anything, userID, 20, 0).Return(follows, int64(1), nil)

		svc := NewFollowService(mockRepo, mockStorage)
		result, err := svc.ListFollows(ctx, userID.String(), 20, 0)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Data, 1)
		assert.Equal(t, targetUser.ID, result.Data[0].User.ID)
		assert.Equal(t, "target_user", result.Data[0].User.Username)
		assert.Equal(t, "Target User", result.Data[0].User.DisplayName)
		assert.Nil(t, result.Data[0].User.Avatar)
		assert.Equal(t, int64(1), result.Pagination.Total)
		assert.Equal(t, 20, result.Pagination.Limit)
		assert.Equal(t, 0, result.Pagination.Offset)
		mockRepo.AssertExpectations(t)
	})

	t.Run("アバター付きのフォローユーザーを取得できる", func(t *testing.T) {
		mockRepo := new(mockFollowRepository)
		mockStorage := new(mockStorageClient)

		avatarID := uuid.New()
		targetUser := model.User{
			ID:          uuid.New(),
			Username:    "avatar_user",
			DisplayName: "Avatar User",
			AvatarID:    &avatarID,
			Avatar: &model.Image{
				ID:   avatarID,
				Path: "https://example.com/avatar.png",
			},
		}
		follows := []model.Follow{
			{
				ID:           uuid.New(),
				UserID:       userID,
				TargetUserID: targetUser.ID,
				CreatedAt:    time.Now(),
				TargetUser:   targetUser,
			},
		}
		mockRepo.On("FindByUserID", mock.Anything, userID, 20, 0).Return(follows, int64(1), nil)

		svc := NewFollowService(mockRepo, mockStorage)
		result, err := svc.ListFollows(ctx, userID.String(), 20, 0)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Data, 1)
		assert.NotNil(t, result.Data[0].User.Avatar)
		assert.Equal(t, avatarID, result.Data[0].User.Avatar.ID)
		assert.Equal(t, "https://example.com/avatar.png", result.Data[0].User.Avatar.URL)
		mockRepo.AssertExpectations(t)
	})

	t.Run("GCS パスのアバターは署名付き URL が生成される", func(t *testing.T) {
		mockRepo := new(mockFollowRepository)
		mockStorage := new(mockStorageClient)

		avatarID := uuid.New()
		targetUser := model.User{
			ID:          uuid.New(),
			Username:    "gcs_user",
			DisplayName: "GCS User",
			AvatarID:    &avatarID,
			Avatar: &model.Image{
				ID:   avatarID,
				Path: "images/avatar.png",
			},
		}
		follows := []model.Follow{
			{
				ID:           uuid.New(),
				UserID:       userID,
				TargetUserID: targetUser.ID,
				CreatedAt:    time.Now(),
				TargetUser:   targetUser,
			},
		}
		mockRepo.On("FindByUserID", mock.Anything, userID, 20, 0).Return(follows, int64(1), nil)
		mockStorage.On("GenerateSignedURL", mock.Anything, "images/avatar.png", mock.Anything).Return("https://storage.example.com/signed-url", nil)

		svc := NewFollowService(mockRepo, mockStorage)
		result, err := svc.ListFollows(ctx, userID.String(), 20, 0)

		assert.NoError(t, err)
		assert.NotNil(t, result.Data[0].User.Avatar)
		assert.Equal(t, "https://storage.example.com/signed-url", result.Data[0].User.Avatar.URL)
		mockRepo.AssertExpectations(t)
		mockStorage.AssertExpectations(t)
	})

	t.Run("空のフォロー一覧を取得できる", func(t *testing.T) {
		mockRepo := new(mockFollowRepository)
		mockStorage := new(mockStorageClient)

		mockRepo.On("FindByUserID", mock.Anything, userID, 20, 0).Return([]model.Follow{}, int64(0), nil)

		svc := NewFollowService(mockRepo, mockStorage)
		result, err := svc.ListFollows(ctx, userID.String(), 20, 0)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result.Data)
		assert.Equal(t, int64(0), result.Pagination.Total)
		mockRepo.AssertExpectations(t)
	})

	t.Run("無効な UUID の場合はエラーを返す", func(t *testing.T) {
		mockRepo := new(mockFollowRepository)
		mockStorage := new(mockStorageClient)

		svc := NewFollowService(mockRepo, mockStorage)
		result, err := svc.ListFollows(ctx, "invalid-uuid", 20, 0)

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("リポジトリがエラーを返すとエラーを返す", func(t *testing.T) {
		mockRepo := new(mockFollowRepository)
		mockStorage := new(mockStorageClient)

		mockRepo.On("FindByUserID", mock.Anything, userID, 20, 0).Return(nil, int64(0), apperror.ErrInternal.WithMessage("Database error"))

		svc := NewFollowService(mockRepo, mockStorage)
		result, err := svc.ListFollows(ctx, userID.String(), 20, 0)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}
