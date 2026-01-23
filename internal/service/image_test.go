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

type mockImageRepository struct {
	mock.Mock
}

func (m *mockImageRepository) Create(ctx context.Context, image *model.Image) error {
	args := m.Called(ctx, image)
	return args.Error(0)
}

func (m *mockImageRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Image, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Image), args.Error(1)
}

func (m *mockImageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockImageRepository) FindOrphaned(ctx context.Context) ([]model.Image, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Image), args.Error(1)
}

func TestUploadImage(t *testing.T) {
	ctx := context.Background()

	t.Run("正常系: png ファイルのアップロードに成功する", func(t *testing.T) {
		mockRepo := new(mockImageRepository)
		mockStorage := new(mockStorageClient)

		svc := NewImageService(mockRepo, mockStorage)

		input := UploadImageInput{
			File:        bytes.NewReader([]byte("fake image data")),
			Filename:    "test.png",
			ContentType: "image/png",
			FileSize:    100,
		}

		mockStorage.On("Upload", mock.Anything, mock.Anything, mock.Anything, "image/png").Return("", nil)
		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Image")).Return(nil)
		mockStorage.On("GenerateSignedURL", mock.Anything, mock.Anything, mock.Anything).Return("https://signed-url.example.com/image.png", nil)

		result, err := svc.UploadImage(ctx, input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "image/png", result.Data.MimeType)
		assert.Equal(t, "test.png", result.Data.Filename)
		assert.Equal(t, 100, result.Data.FileSize)
		assert.Equal(t, "https://signed-url.example.com/image.png", result.Data.URL)
		mockRepo.AssertExpectations(t)
		mockStorage.AssertExpectations(t)
	})

	t.Run("正常系: jpeg ファイルのアップロードに成功する", func(t *testing.T) {
		mockRepo := new(mockImageRepository)
		mockStorage := new(mockStorageClient)

		svc := NewImageService(mockRepo, mockStorage)

		input := UploadImageInput{
			File:        bytes.NewReader([]byte("fake jpeg data")),
			Filename:    "test.jpg",
			ContentType: "image/jpeg",
			FileSize:    200,
		}

		mockStorage.On("Upload", mock.Anything, mock.Anything, mock.Anything, "image/jpeg").Return("", nil)
		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Image")).Return(nil)
		mockStorage.On("GenerateSignedURL", mock.Anything, mock.Anything, mock.Anything).Return("https://signed-url.example.com/image.jpg", nil)

		result, err := svc.UploadImage(ctx, input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "image/jpeg", result.Data.MimeType)
	})

	t.Run("正常系: gif ファイルのアップロードに成功する", func(t *testing.T) {
		mockRepo := new(mockImageRepository)
		mockStorage := new(mockStorageClient)

		svc := NewImageService(mockRepo, mockStorage)

		input := UploadImageInput{
			File:        bytes.NewReader([]byte("fake gif data")),
			Filename:    "test.gif",
			ContentType: "image/gif",
			FileSize:    150,
		}

		mockStorage.On("Upload", mock.Anything, mock.Anything, mock.Anything, "image/gif").Return("", nil)
		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Image")).Return(nil)
		mockStorage.On("GenerateSignedURL", mock.Anything, mock.Anything, mock.Anything).Return("https://signed-url.example.com/image.gif", nil)

		result, err := svc.UploadImage(ctx, input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "image/gif", result.Data.MimeType)
	})

	t.Run("正常系: webp ファイルのアップロードに成功する", func(t *testing.T) {
		mockRepo := new(mockImageRepository)
		mockStorage := new(mockStorageClient)

		svc := NewImageService(mockRepo, mockStorage)

		input := UploadImageInput{
			File:        bytes.NewReader([]byte("fake webp data")),
			Filename:    "test.webp",
			ContentType: "image/webp",
			FileSize:    120,
		}

		mockStorage.On("Upload", mock.Anything, mock.Anything, mock.Anything, "image/webp").Return("", nil)
		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Image")).Return(nil)
		mockStorage.On("GenerateSignedURL", mock.Anything, mock.Anything, mock.Anything).Return("https://signed-url.example.com/image.webp", nil)

		result, err := svc.UploadImage(ctx, input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "image/webp", result.Data.MimeType)
	})

	t.Run("異常系: 無効な MIME タイプの場合エラーを返す", func(t *testing.T) {
		mockRepo := new(mockImageRepository)
		mockStorage := new(mockStorageClient)

		svc := NewImageService(mockRepo, mockStorage)

		input := UploadImageInput{
			File:        bytes.NewReader([]byte("invalid data")),
			Filename:    "test.txt",
			ContentType: "text/plain",
			FileSize:    50,
		}

		result, err := svc.UploadImage(ctx, input)

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.True(t, apperror.IsCode(err, apperror.CodeValidation))
	})

	t.Run("異常系: GCS アップロード失敗時にエラーを返す", func(t *testing.T) {
		mockRepo := new(mockImageRepository)
		mockStorage := new(mockStorageClient)

		svc := NewImageService(mockRepo, mockStorage)

		input := UploadImageInput{
			File:        bytes.NewReader([]byte("fake image data")),
			Filename:    "test.png",
			ContentType: "image/png",
			FileSize:    100,
		}

		mockStorage.On("Upload", mock.Anything, mock.Anything, mock.Anything, "image/png").Return("", errors.New("upload failed"))

		result, err := svc.UploadImage(ctx, input)

		assert.Nil(t, result)
		assert.Error(t, err)
		mockStorage.AssertExpectations(t)
	})

	t.Run("異常系: DB 保存失敗時に GCS からファイルを削除する", func(t *testing.T) {
		mockRepo := new(mockImageRepository)
		mockStorage := new(mockStorageClient)

		svc := NewImageService(mockRepo, mockStorage)

		input := UploadImageInput{
			File:        bytes.NewReader([]byte("fake image data")),
			Filename:    "test.png",
			ContentType: "image/png",
			FileSize:    100,
		}

		mockStorage.On("Upload", mock.Anything, mock.Anything, mock.Anything, "image/png").Return("", nil)
		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Image")).Return(apperror.ErrInternal)
		mockStorage.On("Delete", mock.Anything, mock.Anything).Return(nil)

		result, err := svc.UploadImage(ctx, input)

		assert.Nil(t, result)
		assert.Error(t, err)
		mockStorage.AssertExpectations(t)
	})

	t.Run("異常系: 署名付き URL 生成失敗時にエラーを返す", func(t *testing.T) {
		mockRepo := new(mockImageRepository)
		mockStorage := new(mockStorageClient)

		svc := NewImageService(mockRepo, mockStorage)

		input := UploadImageInput{
			File:        bytes.NewReader([]byte("fake image data")),
			Filename:    "test.png",
			ContentType: "image/png",
			FileSize:    100,
		}

		mockStorage.On("Upload", mock.Anything, mock.Anything, mock.Anything, "image/png").Return("", nil)
		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Image")).Return(nil)
		mockStorage.On("GenerateSignedURL", mock.Anything, mock.Anything, mock.Anything).Return("", errors.New("signed url failed"))

		result, err := svc.UploadImage(ctx, input)

		assert.Nil(t, result)
		assert.Error(t, err)
		mockStorage.AssertExpectations(t)
	})
}

func TestAllowedImageMimeTypes(t *testing.T) {
	t.Run("許可される MIME タイプが正しく定義されている", func(t *testing.T) {
		expected := map[string]string{
			"image/png":  ".png",
			"image/jpeg": ".jpg",
			"image/gif":  ".gif",
			"image/webp": ".webp",
		}

		assert.Equal(t, expected, allowedImageMimeTypes)
	})
}
