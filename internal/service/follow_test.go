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
	"github.com/siropaca/anycast-backend/internal/repository"
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

func (m *mockFollowRepository) Create(ctx context.Context, follow *model.Follow) error {
	args := m.Called(ctx, follow)
	return args.Error(0)
}

func (m *mockFollowRepository) DeleteByUserIDAndTargetUserID(ctx context.Context, userID, targetUserID uuid.UUID) error {
	args := m.Called(ctx, userID, targetUserID)
	return args.Error(0)
}

func (m *mockFollowRepository) ExistsByUserIDAndTargetUserID(ctx context.Context, userID, targetUserID uuid.UUID) (bool, error) {
	args := m.Called(ctx, userID, targetUserID)
	return args.Bool(0), args.Error(1)
}

// UserRepository のモック
type mockUserRepository struct {
	mock.Mock
}

func (m *mockUserRepository) Create(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepository) Update(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *mockUserRepository) FindByIDWithAvatar(ctx context.Context, id uuid.UUID) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *mockUserRepository) FindByUsernameWithAvatar(ctx context.Context, username string) (*model.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *mockUserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *mockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *mockUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	args := m.Called(ctx, username)
	return args.Bool(0), args.Error(1)
}

func (m *mockUserRepository) Search(ctx context.Context, filter repository.SearchUserFilter) ([]model.User, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]model.User), args.Get(1).(int64), args.Error(2)
}

func (m *mockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestNewFollowService(t *testing.T) {
	t.Run("FollowService を作成できる", func(t *testing.T) {
		mockRepo := new(mockFollowRepository)
		mockUserRepo := new(mockUserRepository)
		mockStorage := new(mockStorageClient)
		svc := NewFollowService(mockRepo, mockUserRepo, mockStorage)

		assert.NotNil(t, svc)
	})
}

func TestFollowService_ListFollows(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("フォロー中のユーザー一覧を取得できる", func(t *testing.T) {
		mockRepo := new(mockFollowRepository)
		mockUserRepo := new(mockUserRepository)
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

		svc := NewFollowService(mockRepo, mockUserRepo, mockStorage)
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
		mockUserRepo := new(mockUserRepository)
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

		svc := NewFollowService(mockRepo, mockUserRepo, mockStorage)
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
		mockUserRepo := new(mockUserRepository)
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

		svc := NewFollowService(mockRepo, mockUserRepo, mockStorage)
		result, err := svc.ListFollows(ctx, userID.String(), 20, 0)

		assert.NoError(t, err)
		assert.NotNil(t, result.Data[0].User.Avatar)
		assert.Equal(t, "https://storage.example.com/signed-url", result.Data[0].User.Avatar.URL)
		mockRepo.AssertExpectations(t)
		mockStorage.AssertExpectations(t)
	})

	t.Run("空のフォロー一覧を取得できる", func(t *testing.T) {
		mockRepo := new(mockFollowRepository)
		mockUserRepo := new(mockUserRepository)
		mockStorage := new(mockStorageClient)

		mockRepo.On("FindByUserID", mock.Anything, userID, 20, 0).Return([]model.Follow{}, int64(0), nil)

		svc := NewFollowService(mockRepo, mockUserRepo, mockStorage)
		result, err := svc.ListFollows(ctx, userID.String(), 20, 0)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result.Data)
		assert.Equal(t, int64(0), result.Pagination.Total)
		mockRepo.AssertExpectations(t)
	})

	t.Run("無効な UUID の場合はエラーを返す", func(t *testing.T) {
		mockRepo := new(mockFollowRepository)
		mockUserRepo := new(mockUserRepository)
		mockStorage := new(mockStorageClient)

		svc := NewFollowService(mockRepo, mockUserRepo, mockStorage)
		result, err := svc.ListFollows(ctx, "invalid-uuid", 20, 0)

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("リポジトリがエラーを返すとエラーを返す", func(t *testing.T) {
		mockRepo := new(mockFollowRepository)
		mockUserRepo := new(mockUserRepository)
		mockStorage := new(mockStorageClient)

		mockRepo.On("FindByUserID", mock.Anything, userID, 20, 0).Return(nil, int64(0), apperror.ErrInternal.WithMessage("Database error"))

		svc := NewFollowService(mockRepo, mockUserRepo, mockStorage)
		result, err := svc.ListFollows(ctx, userID.String(), 20, 0)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestFollowService_GetFollowStatus(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	targetUserID := uuid.New()
	targetUsername := "target_user"

	t.Run("フォロー中の場合 following: true を返す", func(t *testing.T) {
		mockRepo := new(mockFollowRepository)
		mockUserRepo := new(mockUserRepository)
		mockStorage := new(mockStorageClient)

		targetUser := &model.User{ID: targetUserID, Username: targetUsername}
		mockUserRepo.On("FindByUsernameWithAvatar", mock.Anything, targetUsername).Return(targetUser, nil)
		mockRepo.On("ExistsByUserIDAndTargetUserID", mock.Anything, userID, targetUserID).Return(true, nil)

		svc := NewFollowService(mockRepo, mockUserRepo, mockStorage)
		result, err := svc.GetFollowStatus(ctx, userID.String(), targetUsername)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.Data.Following)
		mockRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("未フォローの場合 following: false を返す", func(t *testing.T) {
		mockRepo := new(mockFollowRepository)
		mockUserRepo := new(mockUserRepository)
		mockStorage := new(mockStorageClient)

		targetUser := &model.User{ID: targetUserID, Username: targetUsername}
		mockUserRepo.On("FindByUsernameWithAvatar", mock.Anything, targetUsername).Return(targetUser, nil)
		mockRepo.On("ExistsByUserIDAndTargetUserID", mock.Anything, userID, targetUserID).Return(false, nil)

		svc := NewFollowService(mockRepo, mockUserRepo, mockStorage)
		result, err := svc.GetFollowStatus(ctx, userID.String(), targetUsername)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.Data.Following)
		mockRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("無効な UUID の場合はエラーを返す", func(t *testing.T) {
		mockRepo := new(mockFollowRepository)
		mockUserRepo := new(mockUserRepository)
		mockStorage := new(mockStorageClient)

		svc := NewFollowService(mockRepo, mockUserRepo, mockStorage)
		result, err := svc.GetFollowStatus(ctx, "invalid-uuid", targetUsername)

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("存在しないユーザー名の場合はエラーを返す", func(t *testing.T) {
		mockRepo := new(mockFollowRepository)
		mockUserRepo := new(mockUserRepository)
		mockStorage := new(mockStorageClient)

		mockUserRepo.On("FindByUsernameWithAvatar", mock.Anything, "nonexistent").Return(nil, apperror.ErrNotFound.WithMessage("ユーザーが見つかりません"))

		svc := NewFollowService(mockRepo, mockUserRepo, mockStorage)
		result, err := svc.GetFollowStatus(ctx, userID.String(), "nonexistent")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.True(t, apperror.IsCode(err, apperror.CodeNotFound))
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("リポジトリがエラーを返すとエラーを返す", func(t *testing.T) {
		mockRepo := new(mockFollowRepository)
		mockUserRepo := new(mockUserRepository)
		mockStorage := new(mockStorageClient)

		targetUser := &model.User{ID: targetUserID, Username: targetUsername}
		mockUserRepo.On("FindByUsernameWithAvatar", mock.Anything, targetUsername).Return(targetUser, nil)
		mockRepo.On("ExistsByUserIDAndTargetUserID", mock.Anything, userID, targetUserID).Return(false, apperror.ErrInternal.WithMessage("Database error"))

		svc := NewFollowService(mockRepo, mockUserRepo, mockStorage)
		result, err := svc.GetFollowStatus(ctx, userID.String(), targetUsername)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestFollowService_CreateFollow(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	targetUserID := uuid.New()
	targetUsername := "target_user"

	t.Run("フォローを登録できる", func(t *testing.T) {
		mockRepo := new(mockFollowRepository)
		mockUserRepo := new(mockUserRepository)
		mockStorage := new(mockStorageClient)

		targetUser := &model.User{ID: targetUserID, Username: targetUsername}
		mockUserRepo.On("FindByUsernameWithAvatar", mock.Anything, targetUsername).Return(targetUser, nil)
		mockRepo.On("ExistsByUserIDAndTargetUserID", mock.Anything, userID, targetUserID).Return(false, nil)
		mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(f *model.Follow) bool {
			return f.UserID == userID && f.TargetUserID == targetUserID
		})).Return(nil)

		svc := NewFollowService(mockRepo, mockUserRepo, mockStorage)
		result, err := svc.CreateFollow(ctx, userID.String(), targetUsername)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, targetUserID, result.Data.TargetUserID)
		mockRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("自分自身をフォローするとエラーを返す", func(t *testing.T) {
		mockRepo := new(mockFollowRepository)
		mockUserRepo := new(mockUserRepository)
		mockStorage := new(mockStorageClient)

		selfUser := &model.User{ID: userID, Username: "self_user"}
		mockUserRepo.On("FindByUsernameWithAvatar", mock.Anything, "self_user").Return(selfUser, nil)

		svc := NewFollowService(mockRepo, mockUserRepo, mockStorage)
		result, err := svc.CreateFollow(ctx, userID.String(), "self_user")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.True(t, apperror.IsCode(err, apperror.CodeSelfFollowNotAllowed))
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("既にフォロー済みの場合は 409 を返す", func(t *testing.T) {
		mockRepo := new(mockFollowRepository)
		mockUserRepo := new(mockUserRepository)
		mockStorage := new(mockStorageClient)

		targetUser := &model.User{ID: targetUserID, Username: targetUsername}
		mockUserRepo.On("FindByUsernameWithAvatar", mock.Anything, targetUsername).Return(targetUser, nil)
		mockRepo.On("ExistsByUserIDAndTargetUserID", mock.Anything, userID, targetUserID).Return(true, nil)

		svc := NewFollowService(mockRepo, mockUserRepo, mockStorage)
		result, err := svc.CreateFollow(ctx, userID.String(), targetUsername)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.True(t, apperror.IsCode(err, apperror.CodeAlreadyFollowed))
		mockRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("無効な userID の場合はエラーを返す", func(t *testing.T) {
		mockRepo := new(mockFollowRepository)
		mockUserRepo := new(mockUserRepository)
		mockStorage := new(mockStorageClient)

		svc := NewFollowService(mockRepo, mockUserRepo, mockStorage)
		result, err := svc.CreateFollow(ctx, "invalid-uuid", targetUsername)

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("存在しないユーザー名の場合はエラーを返す", func(t *testing.T) {
		mockRepo := new(mockFollowRepository)
		mockUserRepo := new(mockUserRepository)
		mockStorage := new(mockStorageClient)

		mockUserRepo.On("FindByUsernameWithAvatar", mock.Anything, "nonexistent").Return(nil, apperror.ErrNotFound.WithMessage("ユーザーが見つかりません"))

		svc := NewFollowService(mockRepo, mockUserRepo, mockStorage)
		result, err := svc.CreateFollow(ctx, userID.String(), "nonexistent")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.True(t, apperror.IsCode(err, apperror.CodeNotFound))
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("リポジトリの Create がエラーを返すとエラーを返す", func(t *testing.T) {
		mockRepo := new(mockFollowRepository)
		mockUserRepo := new(mockUserRepository)
		mockStorage := new(mockStorageClient)

		targetUser := &model.User{ID: targetUserID, Username: targetUsername}
		mockUserRepo.On("FindByUsernameWithAvatar", mock.Anything, targetUsername).Return(targetUser, nil)
		mockRepo.On("ExistsByUserIDAndTargetUserID", mock.Anything, userID, targetUserID).Return(false, nil)
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(apperror.ErrInternal.WithMessage("Database error"))

		svc := NewFollowService(mockRepo, mockUserRepo, mockStorage)
		result, err := svc.CreateFollow(ctx, userID.String(), targetUsername)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("リポジトリの ExistsByUserIDAndTargetUserID がエラーを返すとエラーを返す", func(t *testing.T) {
		mockRepo := new(mockFollowRepository)
		mockUserRepo := new(mockUserRepository)
		mockStorage := new(mockStorageClient)

		targetUser := &model.User{ID: targetUserID, Username: targetUsername}
		mockUserRepo.On("FindByUsernameWithAvatar", mock.Anything, targetUsername).Return(targetUser, nil)
		mockRepo.On("ExistsByUserIDAndTargetUserID", mock.Anything, userID, targetUserID).Return(false, apperror.ErrInternal.WithMessage("Database error"))

		svc := NewFollowService(mockRepo, mockUserRepo, mockStorage)
		result, err := svc.CreateFollow(ctx, userID.String(), targetUsername)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestFollowService_DeleteFollow(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	targetUserID := uuid.New()
	targetUsername := "target_user"

	t.Run("フォローを解除できる", func(t *testing.T) {
		mockRepo := new(mockFollowRepository)
		mockUserRepo := new(mockUserRepository)
		mockStorage := new(mockStorageClient)

		targetUser := &model.User{ID: targetUserID, Username: targetUsername}
		mockUserRepo.On("FindByUsernameWithAvatar", mock.Anything, targetUsername).Return(targetUser, nil)
		mockRepo.On("DeleteByUserIDAndTargetUserID", mock.Anything, userID, targetUserID).Return(nil)

		svc := NewFollowService(mockRepo, mockUserRepo, mockStorage)
		err := svc.DeleteFollow(ctx, userID.String(), targetUsername)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("存在しないフォローの解除は 404 を返す", func(t *testing.T) {
		mockRepo := new(mockFollowRepository)
		mockUserRepo := new(mockUserRepository)
		mockStorage := new(mockStorageClient)

		targetUser := &model.User{ID: targetUserID, Username: targetUsername}
		mockUserRepo.On("FindByUsernameWithAvatar", mock.Anything, targetUsername).Return(targetUser, nil)
		mockRepo.On("DeleteByUserIDAndTargetUserID", mock.Anything, userID, targetUserID).Return(apperror.ErrNotFound.WithMessage("フォローが見つかりません"))

		svc := NewFollowService(mockRepo, mockUserRepo, mockStorage)
		err := svc.DeleteFollow(ctx, userID.String(), targetUsername)

		assert.Error(t, err)
		assert.True(t, apperror.IsCode(err, apperror.CodeNotFound))
		mockRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("無効な userID の場合はエラーを返す", func(t *testing.T) {
		mockRepo := new(mockFollowRepository)
		mockUserRepo := new(mockUserRepository)
		mockStorage := new(mockStorageClient)

		svc := NewFollowService(mockRepo, mockUserRepo, mockStorage)
		err := svc.DeleteFollow(ctx, "invalid-uuid", targetUsername)

		assert.Error(t, err)
	})

	t.Run("存在しないユーザー名の場合はエラーを返す", func(t *testing.T) {
		mockRepo := new(mockFollowRepository)
		mockUserRepo := new(mockUserRepository)
		mockStorage := new(mockStorageClient)

		mockUserRepo.On("FindByUsernameWithAvatar", mock.Anything, "nonexistent").Return(nil, apperror.ErrNotFound.WithMessage("ユーザーが見つかりません"))

		svc := NewFollowService(mockRepo, mockUserRepo, mockStorage)
		err := svc.DeleteFollow(ctx, userID.String(), "nonexistent")

		assert.Error(t, err)
		assert.True(t, apperror.IsCode(err, apperror.CodeNotFound))
		mockUserRepo.AssertExpectations(t)
	})
}
