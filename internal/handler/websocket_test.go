package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/siropaca/anycast-backend/internal/infrastructure/websocket"
	"github.com/siropaca/anycast-backend/internal/pkg/jwt"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// テスト用のルーターをセットアップする
func setupWebSocketRouter(h *WebSocketHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/ws/jobs", h.HandleJobs)
	return r
}

func TestWebSocketHandler_HandleJobs(t *testing.T) {
	t.Run("トークンが指定されていない場合は 401 を返す", func(t *testing.T) {
		mockTm := new(mockTokenManager)
		hub := websocket.NewHub()

		handler := NewWebSocketHandler(hub, mockTm)
		router := setupWebSocketRouter(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/ws/jobs", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockTm.AssertNotCalled(t, "Validate")
	})

	t.Run("無効なトークンの場合は 401 を返す", func(t *testing.T) {
		mockTm := new(mockTokenManager)
		mockTm.On("Validate", "invalid-token").Return(nil, jwt.ErrInvalidToken)
		hub := websocket.NewHub()

		handler := NewWebSocketHandler(hub, mockTm)
		router := setupWebSocketRouter(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/ws/jobs?token=invalid-token", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockTm.AssertExpectations(t)
	})

	t.Run("有効なトークンで WebSocket 接続を試みる", func(t *testing.T) {
		mockTm := new(mockTokenManager)
		userID := uuid.New().String()
		claims := &jwt.Claims{
			UserID: userID,
		}
		mockTm.On("Validate", "valid-token").Return(claims, nil)
		hub := websocket.NewHub()

		handler := NewWebSocketHandler(hub, mockTm)
		router := setupWebSocketRouter(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/ws/jobs?token=valid-token", http.NoBody)
		router.ServeHTTP(w, req)

		// WebSocket のアップグレードは httptest では完了しないため、
		// トークン検証が通ることだけ確認
		mockTm.AssertExpectations(t)
	})
}

func TestNewWebSocketHandler(t *testing.T) {
	t.Run("WebSocketHandler を作成できる", func(t *testing.T) {
		mockTm := new(mockTokenManager)
		hub := websocket.NewHub()
		handler := NewWebSocketHandler(hub, mockTm)
		assert.NotNil(t, handler)
	})
}
