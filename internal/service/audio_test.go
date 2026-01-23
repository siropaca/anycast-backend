package service

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

type mockAudioRepository struct {
	mock.Mock
}

func (m *mockAudioRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Audio, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Audio), args.Error(1)
}

func (m *mockAudioRepository) Create(ctx context.Context, audio *model.Audio) error {
	args := m.Called(ctx, audio)
	return args.Error(0)
}

func (m *mockAudioRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockAudioRepository) FindOrphaned(ctx context.Context) ([]model.Audio, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Audio), args.Error(1)
}

func TestUploadAudio(t *testing.T) {
	ctx := context.Background()

	t.Run("正常系: mp3 ファイルのアップロードに成功する", func(t *testing.T) {
		mockRepo := new(mockAudioRepository)
		mockStorage := new(mockStorageClient)

		svc := NewAudioService(mockRepo, mockStorage)

		input := UploadAudioInput{
			File:        bytes.NewReader([]byte("fake audio data")),
			Filename:    "test.mp3",
			ContentType: "audio/mpeg",
			FileSize:    100,
		}

		mockStorage.On("Upload", mock.Anything, mock.Anything, mock.Anything, "audio/mpeg").Return("", nil)
		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Audio")).Return(nil)
		mockStorage.On("GenerateSignedURL", mock.Anything, mock.Anything, mock.Anything).Return("https://signed-url.example.com/audio.mp3", nil)

		result, err := svc.UploadAudio(ctx, input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "audio/mpeg", result.Data.MimeType)
		assert.Equal(t, "test.mp3", result.Data.Filename)
		assert.Equal(t, 100, result.Data.FileSize)
		assert.Equal(t, "https://signed-url.example.com/audio.mp3", result.Data.URL)
		mockRepo.AssertExpectations(t)
		mockStorage.AssertExpectations(t)
	})

	t.Run("正常系: wav ファイルのアップロードに成功する", func(t *testing.T) {
		mockRepo := new(mockAudioRepository)
		mockStorage := new(mockStorageClient)

		svc := NewAudioService(mockRepo, mockStorage)

		input := UploadAudioInput{
			File:        bytes.NewReader([]byte("fake wav data")),
			Filename:    "test.wav",
			ContentType: "audio/wav",
			FileSize:    200,
		}

		mockStorage.On("Upload", mock.Anything, mock.Anything, mock.Anything, "audio/wav").Return("", nil)
		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Audio")).Return(nil)
		mockStorage.On("GenerateSignedURL", mock.Anything, mock.Anything, mock.Anything).Return("https://signed-url.example.com/audio.wav", nil)

		result, err := svc.UploadAudio(ctx, input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "audio/wav", result.Data.MimeType)
	})

	t.Run("異常系: 無効な MIME タイプの場合エラーを返す", func(t *testing.T) {
		mockRepo := new(mockAudioRepository)
		mockStorage := new(mockStorageClient)

		svc := NewAudioService(mockRepo, mockStorage)

		input := UploadAudioInput{
			File:        bytes.NewReader([]byte("invalid data")),
			Filename:    "test.txt",
			ContentType: "text/plain",
			FileSize:    50,
		}

		result, err := svc.UploadAudio(ctx, input)

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.True(t, apperror.IsCode(err, apperror.CodeValidation))
	})

	t.Run("異常系: GCS アップロード失敗時にエラーを返す", func(t *testing.T) {
		mockRepo := new(mockAudioRepository)
		mockStorage := new(mockStorageClient)

		svc := NewAudioService(mockRepo, mockStorage)

		input := UploadAudioInput{
			File:        bytes.NewReader([]byte("fake audio data")),
			Filename:    "test.mp3",
			ContentType: "audio/mpeg",
			FileSize:    100,
		}

		mockStorage.On("Upload", mock.Anything, mock.Anything, mock.Anything, "audio/mpeg").Return("", errors.New("upload failed"))

		result, err := svc.UploadAudio(ctx, input)

		assert.Nil(t, result)
		assert.Error(t, err)
		mockStorage.AssertExpectations(t)
	})

	t.Run("異常系: DB 保存失敗時に GCS からファイルを削除する", func(t *testing.T) {
		mockRepo := new(mockAudioRepository)
		mockStorage := new(mockStorageClient)

		svc := NewAudioService(mockRepo, mockStorage)

		input := UploadAudioInput{
			File:        bytes.NewReader([]byte("fake audio data")),
			Filename:    "test.mp3",
			ContentType: "audio/mpeg",
			FileSize:    100,
		}

		mockStorage.On("Upload", mock.Anything, mock.Anything, mock.Anything, "audio/mpeg").Return("", nil)
		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Audio")).Return(apperror.ErrInternal)
		mockStorage.On("Delete", mock.Anything, mock.Anything).Return(nil)

		result, err := svc.UploadAudio(ctx, input)

		assert.Nil(t, result)
		assert.Error(t, err)
		mockStorage.AssertExpectations(t)
	})

	t.Run("異常系: 署名付き URL 生成失敗時にエラーを返す", func(t *testing.T) {
		mockRepo := new(mockAudioRepository)
		mockStorage := new(mockStorageClient)

		svc := NewAudioService(mockRepo, mockStorage)

		input := UploadAudioInput{
			File:        bytes.NewReader([]byte("fake audio data")),
			Filename:    "test.mp3",
			ContentType: "audio/mpeg",
			FileSize:    100,
		}

		mockStorage.On("Upload", mock.Anything, mock.Anything, mock.Anything, "audio/mpeg").Return("", nil)
		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Audio")).Return(nil)
		mockStorage.On("GenerateSignedURL", mock.Anything, mock.Anything, mock.Anything).Return("", errors.New("signed url failed"))

		result, err := svc.UploadAudio(ctx, input)

		assert.Nil(t, result)
		assert.Error(t, err)
		mockStorage.AssertExpectations(t)
	})
}

func TestAllowedAudioMimeTypes(t *testing.T) {
	t.Run("許可される MIME タイプが正しく定義されている", func(t *testing.T) {
		expected := map[string]string{
			"audio/mpeg":  ".mp3",
			"audio/mp3":   ".mp3",
			"audio/wav":   ".wav",
			"audio/wave":  ".wav",
			"audio/ogg":   ".ogg",
			"audio/aac":   ".aac",
			"audio/mp4":   ".m4a",
			"audio/x-m4a": ".m4a",
		}

		assert.Equal(t, expected, allowedAudioMimeTypes)
	})
}
