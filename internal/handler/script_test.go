package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/middleware"
)

// ScriptService のモック
type mockScriptService struct {
	mock.Mock
}

func (m *mockScriptService) GenerateScript(ctx context.Context, userID, channelID, episodeID, prompt string, durationMinutes *int, withEmotion bool) (*response.ScriptLineListResponse, error) {
	args := m.Called(ctx, userID, channelID, episodeID, prompt, durationMinutes, withEmotion)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.ScriptLineListResponse), args.Error(1)
}

// テスト用のルーターをセットアップする
func setupScriptRouter(h *ScriptHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/channels/:channelId/episodes/:episodeId/script/generate", h.GenerateScript)
	return r
}

// 認証済みルーターをセットアップする
func setupAuthenticatedScriptRouter(h *ScriptHandler, userID string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(string(middleware.UserIDKey), userID)
		c.Next()
	})
	r.POST("/channels/:channelId/episodes/:episodeId/script/generate", h.GenerateScript)
	return r
}

func TestScriptHandler_GenerateScript(t *testing.T) {
	userID := uuid.New().String()
	channelID := uuid.New().String()
	episodeID := uuid.New().String()

	t.Run("台本を生成できる", func(t *testing.T) {
		mockSvc := new(mockScriptService)
		lines := []response.ScriptLineResponse{createTestScriptLineResponse()}
		result := &response.ScriptLineListResponse{Data: lines}
		mockSvc.On("GenerateScript", mock.Anything, userID, channelID, episodeID, "AI について語る", (*int)(nil), false).Return(result, nil)

		handler := NewScriptHandler(mockSvc)
		router := setupAuthenticatedScriptRouter(handler, userID)

		body := `{"prompt":"AI について語る"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/channels/"+channelID+"/episodes/"+episodeID+"/script/generate", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp response.ScriptLineListResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Len(t, resp.Data, 1)
		mockSvc.AssertExpectations(t)
	})

	t.Run("durationMinutes を指定して台本を生成できる", func(t *testing.T) {
		mockSvc := new(mockScriptService)
		lines := []response.ScriptLineResponse{createTestScriptLineResponse()}
		result := &response.ScriptLineListResponse{Data: lines}
		duration := 30
		mockSvc.On("GenerateScript", mock.Anything, userID, channelID, episodeID, "AI について語る", &duration, false).Return(result, nil)

		handler := NewScriptHandler(mockSvc)
		router := setupAuthenticatedScriptRouter(handler, userID)

		body := `{"prompt":"AI について語る","durationMinutes":30}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/channels/"+channelID+"/episodes/"+episodeID+"/script/generate", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("withEmotion を指定して台本を生成できる", func(t *testing.T) {
		mockSvc := new(mockScriptService)
		lines := []response.ScriptLineResponse{createTestScriptLineResponse()}
		result := &response.ScriptLineListResponse{Data: lines}
		mockSvc.On("GenerateScript", mock.Anything, userID, channelID, episodeID, "AI について語る", (*int)(nil), true).Return(result, nil)

		handler := NewScriptHandler(mockSvc)
		router := setupAuthenticatedScriptRouter(handler, userID)

		body := `{"prompt":"AI について語る","withEmotion":true}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/channels/"+channelID+"/episodes/"+episodeID+"/script/generate", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("prompt が空の場合は 400 を返す", func(t *testing.T) {
		mockSvc := new(mockScriptService)
		handler := NewScriptHandler(mockSvc)
		router := setupAuthenticatedScriptRouter(handler, userID)

		body := `{"prompt":""}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/channels/"+channelID+"/episodes/"+episodeID+"/script/generate", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("prompt がない場合は 400 を返す", func(t *testing.T) {
		mockSvc := new(mockScriptService)
		handler := NewScriptHandler(mockSvc)
		router := setupAuthenticatedScriptRouter(handler, userID)

		body := `{}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/channels/"+channelID+"/episodes/"+episodeID+"/script/generate", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("不正な JSON の場合は 400 を返す", func(t *testing.T) {
		mockSvc := new(mockScriptService)
		handler := NewScriptHandler(mockSvc)
		router := setupAuthenticatedScriptRouter(handler, userID)

		body := `{"prompt":}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/channels/"+channelID+"/episodes/"+episodeID+"/script/generate", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("チャンネルが見つからない場合は 404 を返す", func(t *testing.T) {
		mockSvc := new(mockScriptService)
		mockSvc.On("GenerateScript", mock.Anything, userID, channelID, episodeID, mock.Anything, mock.Anything, mock.Anything).Return(nil, apperror.ErrNotFound.WithMessage("Channel not found"))

		handler := NewScriptHandler(mockSvc)
		router := setupAuthenticatedScriptRouter(handler, userID)

		body := `{"prompt":"AI について語る"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/channels/"+channelID+"/episodes/"+episodeID+"/script/generate", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("エピソードが見つからない場合は 404 を返す", func(t *testing.T) {
		mockSvc := new(mockScriptService)
		mockSvc.On("GenerateScript", mock.Anything, userID, channelID, episodeID, mock.Anything, mock.Anything, mock.Anything).Return(nil, apperror.ErrNotFound.WithMessage("Episode not found"))

		handler := NewScriptHandler(mockSvc)
		router := setupAuthenticatedScriptRouter(handler, userID)

		body := `{"prompt":"AI について語る"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/channels/"+channelID+"/episodes/"+episodeID+"/script/generate", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("権限がない場合は 403 を返す", func(t *testing.T) {
		mockSvc := new(mockScriptService)
		mockSvc.On("GenerateScript", mock.Anything, userID, channelID, episodeID, mock.Anything, mock.Anything, mock.Anything).Return(nil, apperror.ErrForbidden.WithMessage("You do not have permission"))

		handler := NewScriptHandler(mockSvc)
		router := setupAuthenticatedScriptRouter(handler, userID)

		body := `{"prompt":"AI について語る"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/channels/"+channelID+"/episodes/"+episodeID+"/script/generate", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("台本生成に失敗した場合は 500 を返す", func(t *testing.T) {
		mockSvc := new(mockScriptService)
		mockSvc.On("GenerateScript", mock.Anything, userID, channelID, episodeID, mock.Anything, mock.Anything, mock.Anything).Return(nil, apperror.ErrGenerationFailed.WithMessage("Failed to generate script"))

		handler := NewScriptHandler(mockSvc)
		router := setupAuthenticatedScriptRouter(handler, userID)

		body := `{"prompt":"AI について語る"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/channels/"+channelID+"/episodes/"+episodeID+"/script/generate", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("サービスがエラーを返すとエラーレスポンスを返す", func(t *testing.T) {
		mockSvc := new(mockScriptService)
		mockSvc.On("GenerateScript", mock.Anything, userID, channelID, episodeID, mock.Anything, mock.Anything, mock.Anything).Return(nil, apperror.ErrInternal)

		handler := NewScriptHandler(mockSvc)
		router := setupAuthenticatedScriptRouter(handler, userID)

		body := `{"prompt":"AI について語る"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/channels/"+channelID+"/episodes/"+episodeID+"/script/generate", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("未認証の場合は 401 を返す", func(t *testing.T) {
		mockSvc := new(mockScriptService)
		handler := NewScriptHandler(mockSvc)
		router := setupScriptRouter(handler)

		body := `{"prompt":"AI について語る"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/channels/"+channelID+"/episodes/"+episodeID+"/script/generate", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
