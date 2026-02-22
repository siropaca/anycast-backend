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
	"github.com/siropaca/anycast-backend/internal/pkg/apikey"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// APIKeyRepository のモック
type mockAPIKeyRepository struct {
	mock.Mock
}

func (m *mockAPIKeyRepository) Create(ctx context.Context, ak *model.APIKey) error {
	args := m.Called(ctx, ak)
	return args.Error(0)
}

func (m *mockAPIKeyRepository) FindByKeyHash(ctx context.Context, keyHash string) (*model.APIKey, error) {
	args := m.Called(ctx, keyHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.APIKey), args.Error(1)
}

func (m *mockAPIKeyRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.APIKey, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.APIKey), args.Error(1)
}

func (m *mockAPIKeyRepository) FindByUserIDAndID(ctx context.Context, userID, id uuid.UUID) (*model.APIKey, error) {
	args := m.Called(ctx, userID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.APIKey), args.Error(1)
}

func (m *mockAPIKeyRepository) ExistsByUserIDAndName(ctx context.Context, userID uuid.UUID, name string) (bool, error) {
	args := m.Called(ctx, userID, name)
	return args.Bool(0), args.Error(1)
}

func (m *mockAPIKeyRepository) UpdateLastUsedAt(ctx context.Context, id uuid.UUID, lastUsedAt time.Time) error {
	args := m.Called(ctx, id, lastUsedAt)
	return args.Error(0)
}

func (m *mockAPIKeyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestAPIKeyService_Create(t *testing.T) {
	t.Run("API キーを作成できる", func(t *testing.T) {
		mockRepo := new(mockAPIKeyRepository)
		svc := NewAPIKeyService(mockRepo)

		uid := uuid.MustParse(testUserID)
		mockRepo.On("ExistsByUserIDAndName", mock.Anything, uid, "Test Key").
			Return(false, nil)
		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.APIKey")).
			Return(nil)

		resp, err := svc.Create(context.Background(), testUserID, request.CreateAPIKeyRequest{
			Name: "Test Key",
		})

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "Test Key", resp.Name)
		assert.Contains(t, resp.Key, "ak_")
		assert.Contains(t, resp.Prefix, "ak_")
		mockRepo.AssertExpectations(t)
	})

	t.Run("同名の API キーが存在する場合はエラーを返す", func(t *testing.T) {
		mockRepo := new(mockAPIKeyRepository)
		svc := NewAPIKeyService(mockRepo)

		uid := uuid.MustParse(testUserID)
		mockRepo.On("ExistsByUserIDAndName", mock.Anything, uid, "Duplicate Key").
			Return(true, nil)

		resp, err := svc.Create(context.Background(), testUserID, request.CreateAPIKeyRequest{
			Name: "Duplicate Key",
		})

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.True(t, apperror.IsCode(err, apperror.CodeDuplicateName))
		mockRepo.AssertExpectations(t)
	})

	t.Run("無効な UUID でエラーを返す", func(t *testing.T) {
		mockRepo := new(mockAPIKeyRepository)
		svc := NewAPIKeyService(mockRepo)

		resp, err := svc.Create(context.Background(), "invalid-uuid", request.CreateAPIKeyRequest{
			Name: "Test Key",
		})

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.True(t, apperror.IsCode(err, apperror.CodeValidation))
	})
}

func TestAPIKeyService_List(t *testing.T) {
	t.Run("API キー一覧を取得できる", func(t *testing.T) {
		mockRepo := new(mockAPIKeyRepository)
		svc := NewAPIKeyService(mockRepo)

		uid := uuid.MustParse(testUserID)
		now := time.Now().UTC()
		apiKeys := []model.APIKey{
			{
				ID:        uuid.New(),
				UserID:    uid,
				Name:      "Key 1",
				Prefix:    "ak_abc123...",
				CreatedAt: now,
			},
			{
				ID:         uuid.New(),
				UserID:     uid,
				Name:       "Key 2",
				Prefix:     "ak_def456...",
				LastUsedAt: &now,
				CreatedAt:  now,
			},
		}

		mockRepo.On("FindByUserID", mock.Anything, uid).Return(apiKeys, nil)

		resp, err := svc.List(context.Background(), testUserID)

		assert.NoError(t, err)
		assert.Len(t, resp, 2)
		assert.Equal(t, "Key 1", resp[0].Name)
		assert.Nil(t, resp[0].LastUsedAt)
		assert.Equal(t, "Key 2", resp[1].Name)
		assert.NotNil(t, resp[1].LastUsedAt)
		mockRepo.AssertExpectations(t)
	})
}

func TestAPIKeyService_Delete(t *testing.T) {
	t.Run("API キーを削除できる", func(t *testing.T) {
		mockRepo := new(mockAPIKeyRepository)
		svc := NewAPIKeyService(mockRepo)

		uid := uuid.MustParse(testUserID)
		akID := uuid.New()

		mockRepo.On("FindByUserIDAndID", mock.Anything, uid, akID).
			Return(&model.APIKey{ID: akID, UserID: uid}, nil)
		mockRepo.On("Delete", mock.Anything, akID).
			Return(nil)

		err := svc.Delete(context.Background(), testUserID, akID.String())

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("他ユーザーの API キーは削除できない", func(t *testing.T) {
		mockRepo := new(mockAPIKeyRepository)
		svc := NewAPIKeyService(mockRepo)

		uid := uuid.MustParse(testUserID)
		akID := uuid.New()

		mockRepo.On("FindByUserIDAndID", mock.Anything, uid, akID).
			Return(nil, apperror.ErrNotFound.WithMessage("API キーが見つかりません"))

		err := svc.Delete(context.Background(), testUserID, akID.String())

		assert.Error(t, err)
		assert.True(t, apperror.IsCode(err, apperror.CodeNotFound))
		mockRepo.AssertExpectations(t)
	})
}

func TestAPIKeyService_Authenticate(t *testing.T) {
	t.Run("有効な API キーで認証できる", func(t *testing.T) {
		mockRepo := new(mockAPIKeyRepository)
		svc := NewAPIKeyService(mockRepo)

		plainKey := "ak_0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
		hash := apikey.HashKey(plainKey)
		uid := uuid.MustParse(testUserID)

		mockRepo.On("FindByKeyHash", mock.Anything, hash).
			Return(&model.APIKey{ID: uuid.New(), UserID: uid}, nil)
		// UpdateLastUsedAt は goroutine で非同期実行されるため Maybe で期待を設定
		mockRepo.On("UpdateLastUsedAt", mock.Anything, mock.Anything, mock.Anything).
			Return(nil).Maybe()

		userID, err := svc.Authenticate(context.Background(), plainKey)

		assert.NoError(t, err)
		assert.Equal(t, testUserID, userID)
		// FindByKeyHash の呼び出しのみ検証（UpdateLastUsedAt は非同期のためスキップ）
		mockRepo.AssertCalled(t, "FindByKeyHash", mock.Anything, hash)
	})

	t.Run("無効な API キーでエラーを返す", func(t *testing.T) {
		mockRepo := new(mockAPIKeyRepository)
		svc := NewAPIKeyService(mockRepo)

		plainKey := "ak_invalid"
		hash := apikey.HashKey(plainKey)

		mockRepo.On("FindByKeyHash", mock.Anything, hash).
			Return(nil, apperror.ErrNotFound.WithMessage("API キーが見つかりません"))

		userID, err := svc.Authenticate(context.Background(), plainKey)

		assert.Error(t, err)
		assert.Empty(t, userID)
		assert.True(t, apperror.IsCode(err, apperror.CodeUnauthorized))
		mockRepo.AssertExpectations(t)
	})
}
