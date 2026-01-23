package handler

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/service"
)

// AudioService のモック
type mockAudioService struct {
	mock.Mock
}

func (m *mockAudioService) UploadAudio(ctx context.Context, input service.UploadAudioInput) (*response.AudioUploadDataResponse, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.AudioUploadDataResponse), args.Error(1)
}

// テスト用のルーターをセットアップする
func setupAudioRouter(h *AudioHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/audios", h.UploadAudio)
	return r
}

// テスト用のマルチパートフォームを作成する
func createAudioMultipartForm(t *testing.T, filename, contentType string, content []byte) (body *bytes.Buffer, formContentType string) {
	body = new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	assert.NoError(t, err)

	_, err = io.Copy(part, bytes.NewReader(content))
	assert.NoError(t, err)

	err = writer.Close()
	assert.NoError(t, err)

	return body, writer.FormDataContentType()
}

func TestAudioHandler_UploadAudio(t *testing.T) {
	t.Run("音声ファイルをアップロードできる", func(t *testing.T) {
		mockSvc := new(mockAudioService)
		audioID := uuid.New()
		result := &response.AudioUploadDataResponse{
			Data: response.AudioUploadResponse{
				ID:         audioID,
				MimeType:   "audio/mpeg",
				URL:        "https://example.com/audio.mp3",
				Filename:   "test.mp3",
				FileSize:   1024,
				DurationMs: 5000,
			},
		}
		mockSvc.On("UploadAudio", mock.Anything, mock.AnythingOfType("service.UploadAudioInput")).Return(result, nil)

		handler := NewAudioHandler(mockSvc)
		router := setupAudioRouter(handler)

		body, contentType := createAudioMultipartForm(t, "test.mp3", "audio/mpeg", []byte("test audio data"))

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/audios", body)
		req.Header.Set("Content-Type", contentType)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("ファイルが指定されていない場合は 400 を返す", func(t *testing.T) {
		mockSvc := new(mockAudioService)
		handler := NewAudioHandler(mockSvc)
		router := setupAudioRouter(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/audios", http.NoBody)
		req.Header.Set("Content-Type", "multipart/form-data")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockSvc.AssertNotCalled(t, "UploadAudio")
	})

	t.Run("サービスがバリデーションエラーを返すと 400 を返す", func(t *testing.T) {
		mockSvc := new(mockAudioService)
		mockSvc.On("UploadAudio", mock.Anything, mock.AnythingOfType("service.UploadAudioInput")).Return(nil, apperror.ErrValidation.WithMessage("無効な音声形式です"))

		handler := NewAudioHandler(mockSvc)
		router := setupAudioRouter(handler)

		body, contentType := createAudioMultipartForm(t, "test.txt", "text/plain", []byte("invalid content"))

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/audios", body)
		req.Header.Set("Content-Type", contentType)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("サービスが内部エラーを返すと 500 を返す", func(t *testing.T) {
		mockSvc := new(mockAudioService)
		mockSvc.On("UploadAudio", mock.Anything, mock.AnythingOfType("service.UploadAudioInput")).Return(nil, apperror.ErrInternal)

		handler := NewAudioHandler(mockSvc)
		router := setupAudioRouter(handler)

		body, contentType := createAudioMultipartForm(t, "test.mp3", "audio/mpeg", []byte("test audio data"))

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/audios", body)
		req.Header.Set("Content-Type", contentType)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})
}

func TestNewAudioHandler(t *testing.T) {
	t.Run("AudioHandler を作成できる", func(t *testing.T) {
		mockSvc := new(mockAudioService)
		handler := NewAudioHandler(mockSvc)
		assert.NotNil(t, handler)
	})
}
