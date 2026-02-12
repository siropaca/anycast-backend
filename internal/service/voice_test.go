package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// VoiceRepository のモック
type mockVoiceRepository struct {
	mock.Mock
}

func (m *mockVoiceRepository) FindAll(ctx context.Context, filter repository.VoiceFilter) ([]model.Voice, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Voice), args.Error(1)
}

func (m *mockVoiceRepository) FindByID(ctx context.Context, id string) (*model.Voice, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Voice), args.Error(1)
}

func (m *mockVoiceRepository) FindActiveByID(ctx context.Context, id string) (*model.Voice, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Voice), args.Error(1)
}

// FavoriteVoiceRepository のモック
type mockFavoriteVoiceRepository struct {
	mock.Mock
}

func (m *mockFavoriteVoiceRepository) Create(ctx context.Context, fav *model.FavoriteVoice) error {
	args := m.Called(ctx, fav)
	return args.Error(0)
}

func (m *mockFavoriteVoiceRepository) DeleteByUserIDAndVoiceID(ctx context.Context, userID, voiceID uuid.UUID) error {
	args := m.Called(ctx, userID, voiceID)
	return args.Error(0)
}

func (m *mockFavoriteVoiceRepository) FindVoiceIDsByUserID(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]uuid.UUID), args.Error(1)
}

func (m *mockFavoriteVoiceRepository) ExistsByUserIDAndVoiceID(ctx context.Context, userID, voiceID uuid.UUID) (bool, error) {
	args := m.Called(ctx, userID, voiceID)
	return args.Bool(0), args.Error(1)
}

const testUserID = "8def69af-dae9-4641-a0e5-100107626933"

func TestNewVoiceService(t *testing.T) {
	t.Run("VoiceService を作成できる", func(t *testing.T) {
		mockRepo := new(mockVoiceRepository)
		mockFavRepo := new(mockFavoriteVoiceRepository)
		svc := NewVoiceService(mockRepo, mockFavRepo)

		assert.NotNil(t, svc)
	})
}

func TestVoiceService_ListVoices(t *testing.T) {
	t.Run("ボイス一覧を取得できる", func(t *testing.T) {
		mockRepo := new(mockVoiceRepository)
		mockFavRepo := new(mockFavoriteVoiceRepository)
		voices := []model.Voice{
			{ID: uuid.New(), Name: "Voice 1", Gender: model.GenderMale, Provider: "google", IsActive: true},
			{ID: uuid.New(), Name: "Voice 2", Gender: model.GenderFemale, Provider: "google", IsActive: true},
		}
		filter := repository.VoiceFilter{}
		mockRepo.On("FindAll", mock.Anything, filter).Return(voices, nil)
		mockFavRepo.On("FindVoiceIDsByUserID", mock.Anything, mock.Anything).Return([]uuid.UUID{}, nil)

		svc := NewVoiceService(mockRepo, mockFavRepo)
		result, favIDs, err := svc.ListVoices(context.Background(), testUserID, filter)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Empty(t, favIDs)
		assert.Equal(t, "Voice 1", result[0].Name)
		assert.Equal(t, "Voice 2", result[1].Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("フィルタ条件付きでボイス一覧を取得できる", func(t *testing.T) {
		mockRepo := new(mockVoiceRepository)
		mockFavRepo := new(mockFavoriteVoiceRepository)
		voices := []model.Voice{
			{ID: uuid.New(), Name: "Male Voice", Gender: model.GenderMale, Provider: "google", IsActive: true},
		}
		gender := "male"
		provider := "google"
		filter := repository.VoiceFilter{Gender: &gender, Provider: &provider}
		mockRepo.On("FindAll", mock.Anything, filter).Return(voices, nil)
		mockFavRepo.On("FindVoiceIDsByUserID", mock.Anything, mock.Anything).Return([]uuid.UUID{}, nil)

		svc := NewVoiceService(mockRepo, mockFavRepo)
		result, _, err := svc.ListVoices(context.Background(), testUserID, filter)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, model.GenderMale, result[0].Gender)
		mockRepo.AssertExpectations(t)
	})

	t.Run("空のボイス一覧を取得できる", func(t *testing.T) {
		mockRepo := new(mockVoiceRepository)
		mockFavRepo := new(mockFavoriteVoiceRepository)
		filter := repository.VoiceFilter{}
		mockRepo.On("FindAll", mock.Anything, filter).Return([]model.Voice{}, nil)
		mockFavRepo.On("FindVoiceIDsByUserID", mock.Anything, mock.Anything).Return([]uuid.UUID{}, nil)

		svc := NewVoiceService(mockRepo, mockFavRepo)
		result, _, err := svc.ListVoices(context.Background(), testUserID, filter)

		assert.NoError(t, err)
		assert.Empty(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("リポジトリがエラーを返すとエラーを返す", func(t *testing.T) {
		mockRepo := new(mockVoiceRepository)
		mockFavRepo := new(mockFavoriteVoiceRepository)
		filter := repository.VoiceFilter{}
		mockRepo.On("FindAll", mock.Anything, filter).Return(nil, apperror.ErrInternal.WithMessage("Database error"))

		svc := NewVoiceService(mockRepo, mockFavRepo)
		result, _, err := svc.ListVoices(context.Background(), testUserID, filter)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestVoiceService_GetVoice(t *testing.T) {
	voiceID := uuid.New()

	t.Run("ボイスを取得できる", func(t *testing.T) {
		mockRepo := new(mockVoiceRepository)
		mockFavRepo := new(mockFavoriteVoiceRepository)
		voice := &model.Voice{
			ID:       voiceID,
			Name:     "Test Voice",
			Gender:   model.GenderMale,
			Provider: "google",
			IsActive: true,
		}
		mockRepo.On("FindByID", mock.Anything, voiceID.String()).Return(voice, nil)
		mockFavRepo.On("ExistsByUserIDAndVoiceID", mock.Anything, mock.Anything, mock.Anything).Return(false, nil)

		svc := NewVoiceService(mockRepo, mockFavRepo)
		result, isFav, err := svc.GetVoice(context.Background(), testUserID, voiceID.String())

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, isFav)
		assert.Equal(t, voiceID, result.ID)
		assert.Equal(t, "Test Voice", result.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ボイスが見つからない場合はエラーを返す", func(t *testing.T) {
		mockRepo := new(mockVoiceRepository)
		mockFavRepo := new(mockFavoriteVoiceRepository)
		mockRepo.On("FindByID", mock.Anything, voiceID.String()).Return(nil, apperror.ErrNotFound.WithMessage("Voice not found"))

		svc := NewVoiceService(mockRepo, mockFavRepo)
		result, _, err := svc.GetVoice(context.Background(), testUserID, voiceID.String())

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("リポジトリがエラーを返すとエラーを返す", func(t *testing.T) {
		mockRepo := new(mockVoiceRepository)
		mockFavRepo := new(mockFavoriteVoiceRepository)
		mockRepo.On("FindByID", mock.Anything, voiceID.String()).Return(nil, apperror.ErrInternal.WithMessage("Database error"))

		svc := NewVoiceService(mockRepo, mockFavRepo)
		result, _, err := svc.GetVoice(context.Background(), testUserID, voiceID.String())

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}
