package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

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

func TestVoiceService_ListVoices(t *testing.T) {
	t.Run("ボイス一覧を取得できる", func(t *testing.T) {
		mockRepo := new(mockVoiceRepository)
		voices := []model.Voice{
			{ID: uuid.New(), Provider: "google", Name: "Voice 1"},
			{ID: uuid.New(), Provider: "amazon", Name: "Voice 2"},
		}
		mockRepo.On("FindAll", mock.Anything, repository.VoiceFilter{}).Return(voices, nil)

		svc := NewVoiceService(mockRepo)
		result, err := svc.ListVoices(context.Background(), repository.VoiceFilter{})

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "google", result[0].Provider)
		mockRepo.AssertExpectations(t)
	})

	t.Run("フィルタ条件を渡せる", func(t *testing.T) {
		mockRepo := new(mockVoiceRepository)
		provider := "google"
		filter := repository.VoiceFilter{Provider: &provider}
		voices := []model.Voice{
			{ID: uuid.New(), Provider: "google", Name: "Voice 1"},
		}
		mockRepo.On("FindAll", mock.Anything, filter).Return(voices, nil)

		svc := NewVoiceService(mockRepo)
		result, err := svc.ListVoices(context.Background(), filter)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		mockRepo.AssertExpectations(t)
	})

	t.Run("リポジトリがエラーを返すとエラーを伝播する", func(t *testing.T) {
		mockRepo := new(mockVoiceRepository)
		mockRepo.On("FindAll", mock.Anything, repository.VoiceFilter{}).Return(nil, errors.New("db error"))

		svc := NewVoiceService(mockRepo)
		result, err := svc.ListVoices(context.Background(), repository.VoiceFilter{})

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestVoiceService_GetVoice(t *testing.T) {
	t.Run("ID でボイスを取得できる", func(t *testing.T) {
		mockRepo := new(mockVoiceRepository)
		id := uuid.New()
		voice := &model.Voice{ID: id, Provider: "google", Name: "Test Voice"}
		mockRepo.On("FindByID", mock.Anything, id.String()).Return(voice, nil)

		svc := NewVoiceService(mockRepo)
		result, err := svc.GetVoice(context.Background(), id.String())

		assert.NoError(t, err)
		assert.Equal(t, id, result.ID)
		assert.Equal(t, "google", result.Provider)
		mockRepo.AssertExpectations(t)
	})

	t.Run("リポジトリがエラーを返すとエラーを伝播する", func(t *testing.T) {
		mockRepo := new(mockVoiceRepository)
		id := uuid.New().String()
		mockRepo.On("FindByID", mock.Anything, id).Return(nil, errors.New("not found"))

		svc := NewVoiceService(mockRepo)
		result, err := svc.GetVoice(context.Background(), id)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}
