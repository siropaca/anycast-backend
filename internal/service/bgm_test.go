package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

type mockBgmRepository struct {
	mock.Mock
}

func (m *mockBgmRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Bgm, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Bgm), args.Error(1)
}

func (m *mockBgmRepository) FindByUserID(ctx context.Context, userID uuid.UUID, filter repository.BgmFilter) ([]model.Bgm, int64, error) {
	args := m.Called(ctx, userID, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]model.Bgm), args.Get(1).(int64), args.Error(2)
}

func (m *mockBgmRepository) Create(ctx context.Context, bgm *model.Bgm) error {
	args := m.Called(ctx, bgm)
	return args.Error(0)
}

func (m *mockBgmRepository) Update(ctx context.Context, bgm *model.Bgm) error {
	args := m.Called(ctx, bgm)
	return args.Error(0)
}

func (m *mockBgmRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockBgmRepository) IsUsedInAnyEpisode(ctx context.Context, id uuid.UUID) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *mockBgmRepository) ExistsByUserIDAndName(ctx context.Context, userID uuid.UUID, name string, excludeID *uuid.UUID) (bool, error) {
	args := m.Called(ctx, userID, name, excludeID)
	return args.Bool(0), args.Error(1)
}

type mockSystemBgmRepository struct {
	mock.Mock
}

func (m *mockSystemBgmRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.SystemBgm, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.SystemBgm), args.Error(1)
}

func (m *mockSystemBgmRepository) FindActive(ctx context.Context, filter repository.SystemBgmFilter) ([]model.SystemBgm, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]model.SystemBgm), args.Get(1).(int64), args.Error(2)
}

func (m *mockSystemBgmRepository) CountActive(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func TestListMyBgms(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("正常系: ユーザー BGM のみを取得する", func(t *testing.T) {
		mockBgmRepo := new(mockBgmRepository)
		mockSystemBgmRepo := new(mockSystemBgmRepository)
		mockAudioRepo := new(mockAudioRepository)
		mockStorage := new(mockStorageClient)

		svc := NewBgmService(mockBgmRepo, mockSystemBgmRepo, mockAudioRepo, mockStorage)

		now := time.Now()
		audioID := uuid.New()
		bgmID := uuid.New()

		bgms := []model.Bgm{
			{
				ID:        bgmID,
				UserID:    userID,
				AudioID:   audioID,
				Name:      "Test BGM",
				CreatedAt: now,
				UpdatedAt: now,
				Audio: model.Audio{
					ID:         audioID,
					Path:       "audios/test.mp3",
					DurationMs: 60000,
				},
				Episodes: []model.Episode{},
				Channels: []model.Channel{},
			},
		}

		req := request.ListMyBgmsRequest{
			PaginationRequest: request.PaginationRequest{Limit: 10, Offset: 0},
			IncludeSystem:     false,
		}

		mockBgmRepo.On("FindByUserID", mock.Anything, userID, repository.BgmFilter{Limit: 10, Offset: 0}).Return(bgms, int64(1), nil)
		mockStorage.On("GenerateSignedURL", mock.Anything, "audios/test.mp3", mock.Anything).Return("https://signed-url.example.com/test.mp3", nil)

		result, err := svc.ListMyBgms(ctx, userID.String(), req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Data, 1)
		assert.Equal(t, bgmID, result.Data[0].ID)
		assert.Equal(t, "Test BGM", result.Data[0].Name)
		assert.False(t, result.Data[0].IsSystem)
		assert.Equal(t, int64(1), result.Pagination.Total)
		mockBgmRepo.AssertExpectations(t)
	})

	t.Run("正常系: ユーザー BGM → システム BGM の順で取得する", func(t *testing.T) {
		mockBgmRepo := new(mockBgmRepository)
		mockSystemBgmRepo := new(mockSystemBgmRepository)
		mockAudioRepo := new(mockAudioRepository)
		mockStorage := new(mockStorageClient)

		svc := NewBgmService(mockBgmRepo, mockSystemBgmRepo, mockAudioRepo, mockStorage)

		now := time.Now()
		systemAudioID := uuid.New()
		systemBgmID := uuid.New()
		userAudioID := uuid.New()
		userBgmID := uuid.New()

		systemBgms := []model.SystemBgm{
			{
				ID:        systemBgmID,
				Name:      "System BGM",
				AudioID:   systemAudioID,
				IsActive:  true,
				SortOrder: 1,
				CreatedAt: now,
				UpdatedAt: now,
				Audio: model.Audio{
					ID:         systemAudioID,
					Path:       "audios/system.mp3",
					DurationMs: 30000,
				},
			},
		}

		userBgms := []model.Bgm{
			{
				ID:        userBgmID,
				UserID:    userID,
				AudioID:   userAudioID,
				Name:      "User BGM",
				CreatedAt: now,
				UpdatedAt: now,
				Audio: model.Audio{
					ID:         userAudioID,
					Path:       "audios/user.mp3",
					DurationMs: 60000,
				},
				Episodes: []model.Episode{},
				Channels: []model.Channel{},
			},
		}

		req := request.ListMyBgmsRequest{
			PaginationRequest: request.PaginationRequest{Limit: 10, Offset: 0},
			IncludeSystem:     true,
		}

		mockBgmRepo.On("FindByUserID", mock.Anything, userID, repository.BgmFilter{Limit: 0, Offset: 0}).Return(userBgms, int64(1), nil)
		mockSystemBgmRepo.On("CountActive", mock.Anything).Return(int64(1), nil)
		mockBgmRepo.On("FindByUserID", mock.Anything, userID, repository.BgmFilter{Limit: 10, Offset: 0}).Return(userBgms, int64(1), nil)
		mockStorage.On("GenerateSignedURL", mock.Anything, "audios/user.mp3", mock.Anything).Return("https://signed-url.example.com/user.mp3", nil)
		mockSystemBgmRepo.On("FindActive", mock.Anything, repository.SystemBgmFilter{Limit: 9, Offset: 0}).Return(systemBgms, int64(1), nil)
		mockStorage.On("GenerateSignedURL", mock.Anything, "audios/system.mp3", mock.Anything).Return("https://signed-url.example.com/system.mp3", nil)

		result, err := svc.ListMyBgms(ctx, userID.String(), req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Data, 2)
		assert.False(t, result.Data[0].IsSystem)
		assert.True(t, result.Data[1].IsSystem)
		assert.Equal(t, int64(2), result.Pagination.Total)
	})

	t.Run("異常系: 無効な userID の場合エラーを返す", func(t *testing.T) {
		mockBgmRepo := new(mockBgmRepository)
		mockSystemBgmRepo := new(mockSystemBgmRepository)
		mockAudioRepo := new(mockAudioRepository)
		mockStorage := new(mockStorageClient)

		svc := NewBgmService(mockBgmRepo, mockSystemBgmRepo, mockAudioRepo, mockStorage)

		req := request.ListMyBgmsRequest{
			PaginationRequest: request.PaginationRequest{Limit: 10, Offset: 0},
			IncludeSystem:     false,
		}

		result, err := svc.ListMyBgms(ctx, "invalid-uuid", req)

		assert.Nil(t, result)
		assert.Error(t, err)
	})
}

func TestGetMyBgm(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	bgmID := uuid.New()
	audioID := uuid.New()
	now := time.Now()

	t.Run("正常系: 自分の BGM を取得する", func(t *testing.T) {
		mockBgmRepo := new(mockBgmRepository)
		mockSystemBgmRepo := new(mockSystemBgmRepository)
		mockAudioRepo := new(mockAudioRepository)
		mockStorage := new(mockStorageClient)

		svc := NewBgmService(mockBgmRepo, mockSystemBgmRepo, mockAudioRepo, mockStorage)

		bgm := &model.Bgm{
			ID:        bgmID,
			UserID:    userID,
			AudioID:   audioID,
			Name:      "Test BGM",
			CreatedAt: now,
			UpdatedAt: now,
			Audio: model.Audio{
				ID:         audioID,
				Path:       "audios/test.mp3",
				DurationMs: 60000,
			},
			Episodes: []model.Episode{},
			Channels: []model.Channel{},
		}

		mockBgmRepo.On("FindByID", mock.Anything, bgmID).Return(bgm, nil)
		mockStorage.On("GenerateSignedURL", mock.Anything, "audios/test.mp3", mock.Anything).Return("https://signed-url.example.com/test.mp3", nil)

		result, err := svc.GetMyBgm(ctx, userID.String(), bgmID.String())

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, bgmID, result.Data.ID)
		assert.Equal(t, "Test BGM", result.Data.Name)
	})

	t.Run("異常系: 他のユーザーの BGM を取得しようとした場合エラーを返す", func(t *testing.T) {
		mockBgmRepo := new(mockBgmRepository)
		mockSystemBgmRepo := new(mockSystemBgmRepository)
		mockAudioRepo := new(mockAudioRepository)
		mockStorage := new(mockStorageClient)

		svc := NewBgmService(mockBgmRepo, mockSystemBgmRepo, mockAudioRepo, mockStorage)

		otherUserID := uuid.New()
		bgm := &model.Bgm{
			ID:        bgmID,
			UserID:    otherUserID,
			AudioID:   audioID,
			Name:      "Other User BGM",
			CreatedAt: now,
			UpdatedAt: now,
		}

		mockBgmRepo.On("FindByID", mock.Anything, bgmID).Return(bgm, nil)

		result, err := svc.GetMyBgm(ctx, userID.String(), bgmID.String())

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.True(t, apperror.IsCode(err, apperror.CodeNotFound))
	})

	t.Run("異常系: BGM が存在しない場合エラーを返す", func(t *testing.T) {
		mockBgmRepo := new(mockBgmRepository)
		mockSystemBgmRepo := new(mockSystemBgmRepository)
		mockAudioRepo := new(mockAudioRepository)
		mockStorage := new(mockStorageClient)

		svc := NewBgmService(mockBgmRepo, mockSystemBgmRepo, mockAudioRepo, mockStorage)

		mockBgmRepo.On("FindByID", mock.Anything, bgmID).Return(nil, apperror.ErrNotFound)

		result, err := svc.GetMyBgm(ctx, userID.String(), bgmID.String())

		assert.Nil(t, result)
		assert.Error(t, err)
	})
}

func TestCreateBgm(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	audioID := uuid.New()

	t.Run("正常系: BGM を作成する", func(t *testing.T) {
		mockBgmRepo := new(mockBgmRepository)
		mockSystemBgmRepo := new(mockSystemBgmRepository)
		mockAudioRepo := new(mockAudioRepository)
		mockStorage := new(mockStorageClient)

		svc := NewBgmService(mockBgmRepo, mockSystemBgmRepo, mockAudioRepo, mockStorage)

		audio := &model.Audio{
			ID:         audioID,
			Path:       "audios/test.mp3",
			DurationMs: 60000,
		}

		req := request.CreateBgmRequest{
			Name:    "New BGM",
			AudioID: audioID.String(),
		}

		mockBgmRepo.On("ExistsByUserIDAndName", mock.Anything, userID, "New BGM", (*uuid.UUID)(nil)).Return(false, nil)
		mockAudioRepo.On("FindByID", mock.Anything, audioID).Return(audio, nil)
		mockBgmRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Bgm")).Return(nil)
		mockStorage.On("GenerateSignedURL", mock.Anything, "audios/test.mp3", mock.Anything).Return("https://signed-url.example.com/test.mp3", nil)

		result, err := svc.CreateBgm(ctx, userID.String(), req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "New BGM", result.Data.Name)
		assert.False(t, result.Data.IsSystem)
	})

	t.Run("異常系: 同名の BGM が既に存在する場合エラーを返す", func(t *testing.T) {
		mockBgmRepo := new(mockBgmRepository)
		mockSystemBgmRepo := new(mockSystemBgmRepository)
		mockAudioRepo := new(mockAudioRepository)
		mockStorage := new(mockStorageClient)

		svc := NewBgmService(mockBgmRepo, mockSystemBgmRepo, mockAudioRepo, mockStorage)

		req := request.CreateBgmRequest{
			Name:    "Existing BGM",
			AudioID: audioID.String(),
		}

		mockBgmRepo.On("ExistsByUserIDAndName", mock.Anything, userID, "Existing BGM", (*uuid.UUID)(nil)).Return(true, nil)

		result, err := svc.CreateBgm(ctx, userID.String(), req)

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.True(t, apperror.IsCode(err, apperror.CodeDuplicateName))
	})

	t.Run("異常系: 音声ファイルが存在しない場合エラーを返す", func(t *testing.T) {
		mockBgmRepo := new(mockBgmRepository)
		mockSystemBgmRepo := new(mockSystemBgmRepository)
		mockAudioRepo := new(mockAudioRepository)
		mockStorage := new(mockStorageClient)

		svc := NewBgmService(mockBgmRepo, mockSystemBgmRepo, mockAudioRepo, mockStorage)

		req := request.CreateBgmRequest{
			Name:    "New BGM",
			AudioID: audioID.String(),
		}

		mockBgmRepo.On("ExistsByUserIDAndName", mock.Anything, userID, "New BGM", (*uuid.UUID)(nil)).Return(false, nil)
		mockAudioRepo.On("FindByID", mock.Anything, audioID).Return(nil, apperror.ErrNotFound)

		result, err := svc.CreateBgm(ctx, userID.String(), req)

		assert.Nil(t, result)
		assert.Error(t, err)
	})
}

func TestUpdateMyBgm(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	bgmID := uuid.New()
	audioID := uuid.New()
	now := time.Now()

	t.Run("正常系: BGM の名前を更新する", func(t *testing.T) {
		mockBgmRepo := new(mockBgmRepository)
		mockSystemBgmRepo := new(mockSystemBgmRepository)
		mockAudioRepo := new(mockAudioRepository)
		mockStorage := new(mockStorageClient)

		svc := NewBgmService(mockBgmRepo, mockSystemBgmRepo, mockAudioRepo, mockStorage)

		bgm := &model.Bgm{
			ID:        bgmID,
			UserID:    userID,
			AudioID:   audioID,
			Name:      "Old Name",
			CreatedAt: now,
			UpdatedAt: now,
			Audio: model.Audio{
				ID:         audioID,
				Path:       "audios/test.mp3",
				DurationMs: 60000,
			},
			Episodes: []model.Episode{},
			Channels: []model.Channel{},
		}

		newName := "New Name"
		req := request.UpdateBgmRequest{
			Name: &newName,
		}

		mockBgmRepo.On("FindByID", mock.Anything, bgmID).Return(bgm, nil)
		mockBgmRepo.On("ExistsByUserIDAndName", mock.Anything, userID, "New Name", &bgmID).Return(false, nil)
		mockBgmRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.Bgm")).Return(nil)
		mockStorage.On("GenerateSignedURL", mock.Anything, "audios/test.mp3", mock.Anything).Return("https://signed-url.example.com/test.mp3", nil)

		result, err := svc.UpdateMyBgm(ctx, userID.String(), bgmID.String(), req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "New Name", result.Data.Name)
	})

	t.Run("異常系: 他のユーザーの BGM を更新しようとした場合エラーを返す", func(t *testing.T) {
		mockBgmRepo := new(mockBgmRepository)
		mockSystemBgmRepo := new(mockSystemBgmRepository)
		mockAudioRepo := new(mockAudioRepository)
		mockStorage := new(mockStorageClient)

		svc := NewBgmService(mockBgmRepo, mockSystemBgmRepo, mockAudioRepo, mockStorage)

		otherUserID := uuid.New()
		bgm := &model.Bgm{
			ID:        bgmID,
			UserID:    otherUserID,
			AudioID:   audioID,
			Name:      "Other User BGM",
			CreatedAt: now,
			UpdatedAt: now,
		}

		newName := "New Name"
		req := request.UpdateBgmRequest{
			Name: &newName,
		}

		mockBgmRepo.On("FindByID", mock.Anything, bgmID).Return(bgm, nil)

		result, err := svc.UpdateMyBgm(ctx, userID.String(), bgmID.String(), req)

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.True(t, apperror.IsCode(err, apperror.CodeNotFound))
	})

	t.Run("異常系: 同名の BGM が既に存在する場合エラーを返す", func(t *testing.T) {
		mockBgmRepo := new(mockBgmRepository)
		mockSystemBgmRepo := new(mockSystemBgmRepository)
		mockAudioRepo := new(mockAudioRepository)
		mockStorage := new(mockStorageClient)

		svc := NewBgmService(mockBgmRepo, mockSystemBgmRepo, mockAudioRepo, mockStorage)

		bgm := &model.Bgm{
			ID:        bgmID,
			UserID:    userID,
			AudioID:   audioID,
			Name:      "Old Name",
			CreatedAt: now,
			UpdatedAt: now,
		}

		newName := "Duplicate Name"
		req := request.UpdateBgmRequest{
			Name: &newName,
		}

		mockBgmRepo.On("FindByID", mock.Anything, bgmID).Return(bgm, nil)
		mockBgmRepo.On("ExistsByUserIDAndName", mock.Anything, userID, "Duplicate Name", &bgmID).Return(true, nil)

		result, err := svc.UpdateMyBgm(ctx, userID.String(), bgmID.String(), req)

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.True(t, apperror.IsCode(err, apperror.CodeDuplicateName))
	})
}

func TestDeleteMyBgm(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	bgmID := uuid.New()
	audioID := uuid.New()
	now := time.Now()

	t.Run("正常系: BGM を削除する", func(t *testing.T) {
		mockBgmRepo := new(mockBgmRepository)
		mockSystemBgmRepo := new(mockSystemBgmRepository)
		mockAudioRepo := new(mockAudioRepository)
		mockStorage := new(mockStorageClient)

		svc := NewBgmService(mockBgmRepo, mockSystemBgmRepo, mockAudioRepo, mockStorage)

		bgm := &model.Bgm{
			ID:        bgmID,
			UserID:    userID,
			AudioID:   audioID,
			Name:      "Test BGM",
			CreatedAt: now,
			UpdatedAt: now,
		}

		mockBgmRepo.On("FindByID", mock.Anything, bgmID).Return(bgm, nil)
		mockBgmRepo.On("IsUsedInAnyEpisode", mock.Anything, bgmID).Return(false, nil)
		mockBgmRepo.On("Delete", mock.Anything, bgmID).Return(nil)

		err := svc.DeleteMyBgm(ctx, userID.String(), bgmID.String())

		assert.NoError(t, err)
		mockBgmRepo.AssertExpectations(t)
	})

	t.Run("異常系: 他のユーザーの BGM を削除しようとした場合エラーを返す", func(t *testing.T) {
		mockBgmRepo := new(mockBgmRepository)
		mockSystemBgmRepo := new(mockSystemBgmRepository)
		mockAudioRepo := new(mockAudioRepository)
		mockStorage := new(mockStorageClient)

		svc := NewBgmService(mockBgmRepo, mockSystemBgmRepo, mockAudioRepo, mockStorage)

		otherUserID := uuid.New()
		bgm := &model.Bgm{
			ID:        bgmID,
			UserID:    otherUserID,
			AudioID:   audioID,
			Name:      "Other User BGM",
			CreatedAt: now,
			UpdatedAt: now,
		}

		mockBgmRepo.On("FindByID", mock.Anything, bgmID).Return(bgm, nil)

		err := svc.DeleteMyBgm(ctx, userID.String(), bgmID.String())

		assert.Error(t, err)
		assert.True(t, apperror.IsCode(err, apperror.CodeNotFound))
	})

	t.Run("異常系: エピソードで使用中の場合エラーを返す", func(t *testing.T) {
		mockBgmRepo := new(mockBgmRepository)
		mockSystemBgmRepo := new(mockSystemBgmRepository)
		mockAudioRepo := new(mockAudioRepository)
		mockStorage := new(mockStorageClient)

		svc := NewBgmService(mockBgmRepo, mockSystemBgmRepo, mockAudioRepo, mockStorage)

		bgm := &model.Bgm{
			ID:        bgmID,
			UserID:    userID,
			AudioID:   audioID,
			Name:      "Test BGM",
			CreatedAt: now,
			UpdatedAt: now,
		}

		mockBgmRepo.On("FindByID", mock.Anything, bgmID).Return(bgm, nil)
		mockBgmRepo.On("IsUsedInAnyEpisode", mock.Anything, bgmID).Return(true, nil)

		err := svc.DeleteMyBgm(ctx, userID.String(), bgmID.String())

		assert.Error(t, err)
		assert.True(t, apperror.IsCode(err, apperror.CodeBgmInUse))
	})

	t.Run("異常系: BGM が存在しない場合エラーを返す", func(t *testing.T) {
		mockBgmRepo := new(mockBgmRepository)
		mockSystemBgmRepo := new(mockSystemBgmRepository)
		mockAudioRepo := new(mockAudioRepository)
		mockStorage := new(mockStorageClient)

		svc := NewBgmService(mockBgmRepo, mockSystemBgmRepo, mockAudioRepo, mockStorage)

		mockBgmRepo.On("FindByID", mock.Anything, bgmID).Return(nil, apperror.ErrNotFound)

		err := svc.DeleteMyBgm(ctx, userID.String(), bgmID.String())

		assert.Error(t, err)
	})
}
