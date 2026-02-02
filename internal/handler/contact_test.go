package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

// ContactService のモック
type mockContactService struct {
	mock.Mock
}

func (m *mockContactService) CreateContact(ctx context.Context, input service.CreateContactInput) (*response.ContactDataResponse, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.ContactDataResponse), args.Error(1)
}

// テスト用のルーターをセットアップする（認証あり）
func setupContactRouter(h *ContactHandler, userID string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(string(middleware.UserIDKey), userID)
		c.Next()
	})
	r.POST("/contacts", h.CreateContact)
	return r
}

// テスト用のルーターをセットアップする（認証なし）
func setupContactRouterNoAuth(h *ContactHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/contacts", h.CreateContact)
	return r
}

func TestContactHandler_CreateContact(t *testing.T) {
	userID := uuid.New().String()

	t.Run("全フィールドを送信できる", func(t *testing.T) {
		mockSvc := new(mockContactService)
		contactID := uuid.New()
		now := time.Now()
		userAgent := "Mozilla/5.0"
		result := &response.ContactDataResponse{
			Data: response.ContactResponse{
				ID:        contactID,
				Category:  "general",
				Email:     "test@example.com",
				Name:      "テストユーザー",
				Content:   "お問い合わせ内容です",
				UserAgent: &userAgent,
				CreatedAt: now,
			},
		}
		mockSvc.On("CreateContact", mock.Anything, mock.AnythingOfType("service.CreateContactInput")).Return(result, nil)

		handler := NewContactHandler(mockSvc)
		router := setupContactRouter(handler, userID)

		body, _ := json.Marshal(map[string]string{
			"category":  "general",
			"email":     "test@example.com",
			"name":      "テストユーザー",
			"content":   "お問い合わせ内容です",
			"userAgent": "Mozilla/5.0",
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/contacts", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("必須フィールドのみで送信できる", func(t *testing.T) {
		mockSvc := new(mockContactService)
		contactID := uuid.New()
		now := time.Now()
		result := &response.ContactDataResponse{
			Data: response.ContactResponse{
				ID:        contactID,
				Category:  "bug_report",
				Email:     "test@example.com",
				Name:      "テスト",
				Content:   "バグの報告",
				CreatedAt: now,
			},
		}
		mockSvc.On("CreateContact", mock.Anything, mock.AnythingOfType("service.CreateContactInput")).Return(result, nil)

		handler := NewContactHandler(mockSvc)
		router := setupContactRouter(handler, userID)

		body, _ := json.Marshal(map[string]string{
			"category": "bug_report",
			"email":    "test@example.com",
			"name":     "テスト",
			"content":  "バグの報告",
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/contacts", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("未認証でも送信できる", func(t *testing.T) {
		mockSvc := new(mockContactService)
		contactID := uuid.New()
		now := time.Now()
		result := &response.ContactDataResponse{
			Data: response.ContactResponse{
				ID:        contactID,
				Category:  "general",
				Email:     "guest@example.com",
				Name:      "ゲスト",
				Content:   "ゲストからの問い合わせ",
				CreatedAt: now,
			},
		}
		mockSvc.On("CreateContact", mock.Anything, mock.AnythingOfType("service.CreateContactInput")).Return(result, nil)

		handler := NewContactHandler(mockSvc)
		router := setupContactRouterNoAuth(handler)

		body, _ := json.Marshal(map[string]string{
			"category": "general",
			"email":    "guest@example.com",
			"name":     "ゲスト",
			"content":  "ゲストからの問い合わせ",
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/contacts", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("各カテゴリで送信できる", func(t *testing.T) {
		categories := []string{"general", "bug_report", "feature_request", "other"}

		for _, category := range categories {
			t.Run(category, func(t *testing.T) {
				mockSvc := new(mockContactService)
				contactID := uuid.New()
				now := time.Now()
				result := &response.ContactDataResponse{
					Data: response.ContactResponse{
						ID:        contactID,
						Category:  category,
						Email:     "test@example.com",
						Name:      "テスト",
						Content:   "テスト内容",
						CreatedAt: now,
					},
				}
				mockSvc.On("CreateContact", mock.Anything, mock.AnythingOfType("service.CreateContactInput")).Return(result, nil)

				handler := NewContactHandler(mockSvc)
				router := setupContactRouter(handler, userID)

				body, _ := json.Marshal(map[string]string{
					"category": category,
					"email":    "test@example.com",
					"name":     "テスト",
					"content":  "テスト内容",
				})

				w := httptest.NewRecorder()
				req := httptest.NewRequest("POST", "/contacts", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusCreated, w.Code)
				mockSvc.AssertExpectations(t)
			})
		}
	})

	t.Run("カテゴリが空の場合は 400 を返す", func(t *testing.T) {
		mockSvc := new(mockContactService)
		handler := NewContactHandler(mockSvc)
		router := setupContactRouter(handler, userID)

		body, _ := json.Marshal(map[string]string{
			"category": "",
			"email":    "test@example.com",
			"name":     "テスト",
			"content":  "テスト内容",
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/contacts", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockSvc.AssertNotCalled(t, "CreateContact")
	})

	t.Run("無効なカテゴリの場合は 400 を返す", func(t *testing.T) {
		mockSvc := new(mockContactService)
		handler := NewContactHandler(mockSvc)
		router := setupContactRouter(handler, userID)

		body, _ := json.Marshal(map[string]string{
			"category": "invalid",
			"email":    "test@example.com",
			"name":     "テスト",
			"content":  "テスト内容",
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/contacts", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockSvc.AssertNotCalled(t, "CreateContact")
	})

	t.Run("メールアドレスが空の場合は 400 を返す", func(t *testing.T) {
		mockSvc := new(mockContactService)
		handler := NewContactHandler(mockSvc)
		router := setupContactRouter(handler, userID)

		body, _ := json.Marshal(map[string]string{
			"category": "general",
			"email":    "",
			"name":     "テスト",
			"content":  "テスト内容",
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/contacts", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockSvc.AssertNotCalled(t, "CreateContact")
	})

	t.Run("メールアドレスが不正な場合は 400 を返す", func(t *testing.T) {
		mockSvc := new(mockContactService)
		handler := NewContactHandler(mockSvc)
		router := setupContactRouter(handler, userID)

		body, _ := json.Marshal(map[string]string{
			"category": "general",
			"email":    "invalid-email",
			"name":     "テスト",
			"content":  "テスト内容",
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/contacts", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockSvc.AssertNotCalled(t, "CreateContact")
	})

	t.Run("名前が空の場合は 400 を返す", func(t *testing.T) {
		mockSvc := new(mockContactService)
		handler := NewContactHandler(mockSvc)
		router := setupContactRouter(handler, userID)

		body, _ := json.Marshal(map[string]string{
			"category": "general",
			"email":    "test@example.com",
			"name":     "",
			"content":  "テスト内容",
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/contacts", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockSvc.AssertNotCalled(t, "CreateContact")
	})

	t.Run("コンテンツが空の場合は 400 を返す", func(t *testing.T) {
		mockSvc := new(mockContactService)
		handler := NewContactHandler(mockSvc)
		router := setupContactRouter(handler, userID)

		body, _ := json.Marshal(map[string]string{
			"category": "general",
			"email":    "test@example.com",
			"name":     "テスト",
			"content":  "",
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/contacts", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockSvc.AssertNotCalled(t, "CreateContact")
	})

	t.Run("コンテンツが5000文字を超える場合は 400 を返す", func(t *testing.T) {
		mockSvc := new(mockContactService)
		handler := NewContactHandler(mockSvc)
		router := setupContactRouter(handler, userID)

		longContent := strings.Repeat("a", 5001)
		body, _ := json.Marshal(map[string]string{
			"category": "general",
			"email":    "test@example.com",
			"name":     "テスト",
			"content":  longContent,
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/contacts", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockSvc.AssertNotCalled(t, "CreateContact")
	})

	t.Run("サービスがエラーを返すと適切なステータスコードを返す", func(t *testing.T) {
		mockSvc := new(mockContactService)
		mockSvc.On("CreateContact", mock.Anything, mock.AnythingOfType("service.CreateContactInput")).Return(nil, apperror.ErrInternal)

		handler := NewContactHandler(mockSvc)
		router := setupContactRouter(handler, userID)

		body, _ := json.Marshal(map[string]string{
			"category": "general",
			"email":    "test@example.com",
			"name":     "テスト",
			"content":  "テスト内容",
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/contacts", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})
}

func TestNewContactHandler(t *testing.T) {
	t.Run("ContactHandler を作成できる", func(t *testing.T) {
		mockSvc := new(mockContactService)
		handler := NewContactHandler(mockSvc)
		assert.NotNil(t, handler)
	})
}
