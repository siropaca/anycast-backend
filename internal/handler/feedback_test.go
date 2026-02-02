package handler

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/middleware"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/service"
)

// FeedbackService のモック
type mockFeedbackService struct {
	mock.Mock
}

func (m *mockFeedbackService) CreateFeedback(ctx context.Context, userID string, input service.CreateFeedbackInput) (*response.FeedbackDataResponse, error) {
	args := m.Called(ctx, userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.FeedbackDataResponse), args.Error(1)
}

// テスト用のルーターをセットアップする
func setupFeedbackRouter(h *FeedbackHandler, userID string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(string(middleware.UserIDKey), userID)
		c.Next()
	})
	r.POST("/feedbacks", h.CreateFeedback)
	return r
}

// テスト用のマルチパートフォームを作成する（フィードバック用）
func createFeedbackMultipartForm(t *testing.T, content string, screenshot *struct {
	filename    string
	contentType string
	data        []byte
}, pageURL, userAgent *string) (body *bytes.Buffer, formContentType string) {
	body = new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	// content フィールド
	err := writer.WriteField("content", content)
	assert.NoError(t, err)

	// pageUrl フィールド（任意）
	if pageURL != nil {
		err = writer.WriteField("pageUrl", *pageURL)
		assert.NoError(t, err)
	}

	// userAgent フィールド（任意）
	if userAgent != nil {
		err = writer.WriteField("userAgent", *userAgent)
		assert.NoError(t, err)
	}

	// screenshot ファイル（任意）- Content-Type ヘッダー付き
	if screenshot != nil {
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", `form-data; name="screenshot"; filename="`+screenshot.filename+`"`)
		h.Set("Content-Type", screenshot.contentType)
		part, err := writer.CreatePart(h)
		assert.NoError(t, err)
		_, err = io.Copy(part, bytes.NewReader(screenshot.data))
		assert.NoError(t, err)
	}

	err = writer.Close()
	assert.NoError(t, err)

	return body, writer.FormDataContentType()
}

func TestFeedbackHandler_CreateFeedback(t *testing.T) {
	userID := uuid.New().String()

	t.Run("フィードバックを送信できる", func(t *testing.T) {
		mockSvc := new(mockFeedbackService)
		feedbackID := uuid.New()
		now := time.Now()
		result := &response.FeedbackDataResponse{
			Data: response.FeedbackResponse{
				ID:        feedbackID,
				Content:   "テストフィードバック",
				CreatedAt: now,
			},
		}
		mockSvc.On("CreateFeedback", mock.Anything, userID, mock.AnythingOfType("service.CreateFeedbackInput")).Return(result, nil)

		handler := NewFeedbackHandler(mockSvc)
		router := setupFeedbackRouter(handler, userID)

		body, contentType := createFeedbackMultipartForm(t, "テストフィードバック", nil, nil, nil)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/feedbacks", body)
		req.Header.Set("Content-Type", contentType)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("スクリーンショット付きでフィードバックを送信できる", func(t *testing.T) {
		mockSvc := new(mockFeedbackService)
		feedbackID := uuid.New()
		screenshotID := uuid.New()
		now := time.Now()
		result := &response.FeedbackDataResponse{
			Data: response.FeedbackResponse{
				ID:      feedbackID,
				Content: "バグ報告",
				Screenshot: &response.ArtworkResponse{
					ID:  screenshotID,
					URL: "https://example.com/screenshot.png",
				},
				CreatedAt: now,
			},
		}
		mockSvc.On("CreateFeedback", mock.Anything, userID, mock.AnythingOfType("service.CreateFeedbackInput")).Return(result, nil)

		handler := NewFeedbackHandler(mockSvc)
		router := setupFeedbackRouter(handler, userID)

		screenshot := &struct {
			filename    string
			contentType string
			data        []byte
		}{
			filename:    "screenshot.png",
			contentType: "image/png",
			data:        []byte("fake png data"),
		}
		body, contentType := createFeedbackMultipartForm(t, "バグ報告", screenshot, nil, nil)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/feedbacks", body)
		req.Header.Set("Content-Type", contentType)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("メタデータ付きでフィードバックを送信できる", func(t *testing.T) {
		mockSvc := new(mockFeedbackService)
		feedbackID := uuid.New()
		now := time.Now()
		pageURL := "https://app.example.com/channels/123"
		userAgent := "Mozilla/5.0"
		result := &response.FeedbackDataResponse{
			Data: response.FeedbackResponse{
				ID:        feedbackID,
				Content:   "機能リクエスト",
				PageURL:   &pageURL,
				UserAgent: &userAgent,
				CreatedAt: now,
			},
		}
		mockSvc.On("CreateFeedback", mock.Anything, userID, mock.AnythingOfType("service.CreateFeedbackInput")).Return(result, nil)

		handler := NewFeedbackHandler(mockSvc)
		router := setupFeedbackRouter(handler, userID)

		body, contentType := createFeedbackMultipartForm(t, "機能リクエスト", nil, &pageURL, &userAgent)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/feedbacks", body)
		req.Header.Set("Content-Type", contentType)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("フィードバック内容が空の場合は 400 を返す", func(t *testing.T) {
		mockSvc := new(mockFeedbackService)
		handler := NewFeedbackHandler(mockSvc)
		router := setupFeedbackRouter(handler, userID)

		body, contentType := createFeedbackMultipartForm(t, "", nil, nil, nil)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/feedbacks", body)
		req.Header.Set("Content-Type", contentType)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockSvc.AssertNotCalled(t, "CreateFeedback")
	})

	t.Run("フィードバック内容が5000文字を超える場合は 400 を返す", func(t *testing.T) {
		mockSvc := new(mockFeedbackService)
		handler := NewFeedbackHandler(mockSvc)
		router := setupFeedbackRouter(handler, userID)

		// 5001文字のコンテンツ
		longContent := string(make([]byte, 5001))
		for i := range longContent {
			longContent = longContent[:i] + "あ" + longContent[i+1:]
		}
		body, contentType := createFeedbackMultipartForm(t, longContent, nil, nil, nil)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/feedbacks", body)
		req.Header.Set("Content-Type", contentType)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockSvc.AssertNotCalled(t, "CreateFeedback")
	})

	t.Run("無効な画像形式の場合は 400 を返す", func(t *testing.T) {
		mockSvc := new(mockFeedbackService)
		handler := NewFeedbackHandler(mockSvc)
		router := setupFeedbackRouter(handler, userID)

		// Content-Type が text/plain のファイル
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("content", "テストフィードバック")

		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", `form-data; name="screenshot"; filename="test.txt"`)
		h.Set("Content-Type", "text/plain")
		part, _ := writer.CreatePart(h)
		_, _ = part.Write([]byte("invalid file"))
		_ = writer.Close()

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/feedbacks", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockSvc.AssertNotCalled(t, "CreateFeedback")
	})

	t.Run("サービスがエラーを返すと適切なステータスコードを返す", func(t *testing.T) {
		mockSvc := new(mockFeedbackService)
		mockSvc.On("CreateFeedback", mock.Anything, userID, mock.AnythingOfType("service.CreateFeedbackInput")).Return(nil, apperror.ErrInternal)

		handler := NewFeedbackHandler(mockSvc)
		router := setupFeedbackRouter(handler, userID)

		body, contentType := createFeedbackMultipartForm(t, "テストフィードバック", nil, nil, nil)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/feedbacks", body)
		req.Header.Set("Content-Type", contentType)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})
}

func TestNewFeedbackHandler(t *testing.T) {
	t.Run("FeedbackHandler を作成できる", func(t *testing.T) {
		mockSvc := new(mockFeedbackService)
		handler := NewFeedbackHandler(mockSvc)
		assert.NotNil(t, handler)
	})
}

func TestIsValidImageType(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		expected    bool
	}{
		{"PNG は有効", "image/png", true},
		{"JPEG は有効", "image/jpeg", true},
		{"WebP は有効", "image/webp", true},
		{"GIF は無効", "image/gif", false},
		{"テキストは無効", "text/plain", false},
		{"空は無効", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidImageType(tt.contentType)
			assert.Equal(t, tt.expected, result)
		})
	}
}
