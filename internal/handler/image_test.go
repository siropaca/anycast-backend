package handler

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/middleware"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/service"
)

// ImageService のモック
type mockImageService struct {
	mock.Mock
}

func (m *mockImageService) UploadImage(ctx context.Context, input service.UploadImageInput) (*response.ImageUploadDataResponse, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.ImageUploadDataResponse), args.Error(1)
}

func (m *mockImageService) GenerateImage(ctx context.Context, prompt string) (*response.ImageUploadDataResponse, error) {
	args := m.Called(ctx, prompt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.ImageUploadDataResponse), args.Error(1)
}

// テスト用のルーターをセットアップする
func setupImageRouter(h *ImageHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/images", h.UploadImage)
	return r
}

// テスト用の画像生成ルーターをセットアップする
func setupImageGenerateRouter(h *ImageHandler, userID string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	authMiddleware := func(c *gin.Context) {
		c.Set(string(middleware.UserIDKey), userID)
		c.Next()
	}

	r.POST("/images/generate", authMiddleware, h.GenerateImage)

	// 認証なしルート
	r.POST("/no-auth/images/generate", h.GenerateImage)

	return r
}

// テスト用のマルチパートフォームを作成する
func createImageMultipartForm(t *testing.T, filename, contentType string, content []byte) (body *bytes.Buffer, formContentType string) {
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

func TestImageHandler_UploadImage(t *testing.T) {
	t.Run("画像ファイルをアップロードできる", func(t *testing.T) {
		mockSvc := new(mockImageService)
		imageID := uuid.New()
		result := &response.ImageUploadDataResponse{
			Data: response.ImageUploadResponse{
				ID:       imageID,
				MimeType: "image/png",
				URL:      "https://example.com/image.png",
				Filename: "test.png",
				FileSize: 2048,
			},
		}
		mockSvc.On("UploadImage", mock.Anything, mock.AnythingOfType("service.UploadImageInput")).Return(result, nil)

		handler := NewImageHandler(mockSvc)
		router := setupImageRouter(handler)

		body, contentType := createImageMultipartForm(t, "test.png", "image/png", []byte("test image data"))

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/images", body)
		req.Header.Set("Content-Type", contentType)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("ファイルが指定されていない場合は 400 を返す", func(t *testing.T) {
		mockSvc := new(mockImageService)
		handler := NewImageHandler(mockSvc)
		router := setupImageRouter(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/images", http.NoBody)
		req.Header.Set("Content-Type", "multipart/form-data")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockSvc.AssertNotCalled(t, "UploadImage")
	})

	t.Run("サービスがバリデーションエラーを返すと 400 を返す", func(t *testing.T) {
		mockSvc := new(mockImageService)
		mockSvc.On("UploadImage", mock.Anything, mock.AnythingOfType("service.UploadImageInput")).Return(nil, apperror.ErrValidation.WithMessage("無効な画像形式です"))

		handler := NewImageHandler(mockSvc)
		router := setupImageRouter(handler)

		body, contentType := createImageMultipartForm(t, "test.txt", "text/plain", []byte("invalid content"))

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/images", body)
		req.Header.Set("Content-Type", contentType)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("サービスが内部エラーを返すと 500 を返す", func(t *testing.T) {
		mockSvc := new(mockImageService)
		mockSvc.On("UploadImage", mock.Anything, mock.AnythingOfType("service.UploadImageInput")).Return(nil, apperror.ErrInternal)

		handler := NewImageHandler(mockSvc)
		router := setupImageRouter(handler)

		body, contentType := createImageMultipartForm(t, "test.png", "image/png", []byte("test image data"))

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/images", body)
		req.Header.Set("Content-Type", contentType)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})
}

func TestNewImageHandler(t *testing.T) {
	t.Run("ImageHandler を作成できる", func(t *testing.T) {
		mockSvc := new(mockImageService)
		handler := NewImageHandler(mockSvc)
		assert.NotNil(t, handler)
	})
}

func TestImageHandler_GenerateImage(t *testing.T) {
	t.Run("プロンプト指定で画像を生成できる", func(t *testing.T) {
		mockSvc := new(mockImageService)
		imageID := uuid.New()
		result := &response.ImageUploadDataResponse{
			Data: response.ImageUploadResponse{
				ID:       imageID,
				MimeType: "image/png",
				URL:      "https://example.com/generated.png",
				Filename: "generated_12345678.png",
				FileSize: 1234567,
			},
		}
		mockSvc.On("GenerateImage", mock.Anything, "テスト用プロンプト").Return(result, nil)

		handler := NewImageHandler(mockSvc)
		router := setupImageGenerateRouter(handler, "user-123")

		body := `{"prompt":"テスト用プロンプト"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/images/generate", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("未認証の場合は 401 を返す", func(t *testing.T) {
		mockSvc := new(mockImageService)
		handler := NewImageHandler(mockSvc)
		router := setupImageGenerateRouter(handler, "user-123")

		body := `{"prompt":"テスト"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/no-auth/images/generate", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockSvc.AssertNotCalled(t, "GenerateImage")
	})

	t.Run("prompt が空の場合は 400 を返す", func(t *testing.T) {
		mockSvc := new(mockImageService)
		handler := NewImageHandler(mockSvc)
		router := setupImageGenerateRouter(handler, "user-123")

		body := `{"prompt":""}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/images/generate", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockSvc.AssertNotCalled(t, "GenerateImage")
	})

	t.Run("prompt が 1000 文字超の場合は 400 を返す", func(t *testing.T) {
		mockSvc := new(mockImageService)
		handler := NewImageHandler(mockSvc)
		router := setupImageGenerateRouter(handler, "user-123")

		longPrompt := strings.Repeat("あ", 1001)
		body := `{"prompt":"` + longPrompt + `"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/images/generate", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockSvc.AssertNotCalled(t, "GenerateImage")
	})

	t.Run("サービスがエラーを返すと 500 を返す", func(t *testing.T) {
		mockSvc := new(mockImageService)
		mockSvc.On("GenerateImage", mock.Anything, "テスト").
			Return(nil, apperror.ErrGenerationFailed.WithMessage("画像生成に失敗しました"))

		handler := NewImageHandler(mockSvc)
		router := setupImageGenerateRouter(handler, "user-123")

		body := `{"prompt":"テスト"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/images/generate", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})
}
