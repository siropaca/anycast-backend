package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/model"
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

func TestNewVoiceService(t *testing.T) {
	t.Run("VoiceService を作成できる", func(t *testing.T) {
		mockRepo := new(mockVoiceRepository)
		svc := NewVoiceService(mockRepo)

		assert.NotNil(t, svc)
	})
}

func TestVoiceService_ListVoices(t *testing.T) {
	t.Run("ボイス一覧を取得できる", func(t *testing.T) {
		mockRepo := new(mockVoiceRepository)
		voices := []model.Voice{
			{ID: uuid.New(), Name: "Voice 1", Gender: model.GenderMale, Provider: "google", IsActive: true},
			{ID: uuid.New(), Name: "Voice 2", Gender: model.GenderFemale, Provider: "google", IsActive: true},
		}
		filter := repository.VoiceFilter{}
		mockRepo.On("FindAll", mock.Anything, filter).Return(voices, nil)

		svc := NewVoiceService(mockRepo)
		result, err := svc.ListVoices(context.Background(), filter)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "Voice 1", result[0].Name)
		assert.Equal(t, "Voice 2", result[1].Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("フィルタ条件付きでボイス一覧を取得できる", func(t *testing.T) {
		mockRepo := new(mockVoiceRepository)
		voices := []model.Voice{
			{ID: uuid.New(), Name: "Male Voice", Gender: model.GenderMale, Provider: "google", IsActive: true},
		}
		gender := "male"
		provider := "google"
		filter := repository.VoiceFilter{Gender: &gender, Provider: &provider}
		mockRepo.On("FindAll", mock.Anything, filter).Return(voices, nil)

		svc := NewVoiceService(mockRepo)
		result, err := svc.ListVoices(context.Background(), filter)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, model.GenderMale, result[0].Gender)
		mockRepo.AssertExpectations(t)
	})

	t.Run("空のボイス一覧を取得できる", func(t *testing.T) {
		mockRepo := new(mockVoiceRepository)
		filter := repository.VoiceFilter{}
		mockRepo.On("FindAll", mock.Anything, filter).Return([]model.Voice{}, nil)

		svc := NewVoiceService(mockRepo)
		result, err := svc.ListVoices(context.Background(), filter)

		assert.NoError(t, err)
		assert.Empty(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("リポジトリがエラーを返すとエラーを返す", func(t *testing.T) {
		mockRepo := new(mockVoiceRepository)
		filter := repository.VoiceFilter{}
		mockRepo.On("FindAll", mock.Anything, filter).Return(nil, apperror.ErrInternal.WithMessage("Database error"))

		svc := NewVoiceService(mockRepo)
		result, err := svc.ListVoices(context.Background(), filter)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestVoiceService_GetVoice(t *testing.T) {
	voiceID := uuid.New()

	t.Run("ボイスを取得できる", func(t *testing.T) {
		mockRepo := new(mockVoiceRepository)
		voice := &model.Voice{
			ID:       voiceID,
			Name:     "Test Voice",
			Gender:   model.GenderMale,
			Provider: "google",
			IsActive: true,
		}
		mockRepo.On("FindByID", mock.Anything, voiceID.String()).Return(voice, nil)

		svc := NewVoiceService(mockRepo)
		result, err := svc.GetVoice(context.Background(), voiceID.String())

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, voiceID, result.ID)
		assert.Equal(t, "Test Voice", result.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("ボイスが見つからない場合はエラーを返す", func(t *testing.T) {
		mockRepo := new(mockVoiceRepository)
		mockRepo.On("FindByID", mock.Anything, voiceID.String()).Return(nil, apperror.ErrNotFound.WithMessage("Voice not found"))

		svc := NewVoiceService(mockRepo)
		result, err := svc.GetVoice(context.Background(), voiceID.String())

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("リポジトリがエラーを返すとエラーを返す", func(t *testing.T) {
		mockRepo := new(mockVoiceRepository)
		mockRepo.On("FindByID", mock.Anything, voiceID.String()).Return(nil, apperror.ErrInternal.WithMessage("Database error"))

		svc := NewVoiceService(mockRepo)
		result, err := svc.GetVoice(context.Background(), voiceID.String())

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}
