package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	gorillaWs "github.com/gorilla/websocket"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/infrastructure/websocket"
	"github.com/siropaca/anycast-backend/internal/pkg/jwt"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
)

var upgrader = gorillaWs.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: 本番環境では適切なオリジンチェックを行う
		return true
	},
}

// WebSocketHandler は WebSocket 接続を処理するハンドラー
type WebSocketHandler struct {
	hub          *websocket.Hub
	tokenManager jwt.TokenManager
}

// NewWebSocketHandler は WebSocketHandler を作成する
func NewWebSocketHandler(hub *websocket.Hub, tm jwt.TokenManager) *WebSocketHandler {
	return &WebSocketHandler{
		hub:          hub,
		tokenManager: tm,
	}
}

// HandleJobs godoc
// @Summary ジョブの WebSocket 接続
// @Description リアルタイムで音声生成・台本生成ジョブの進捗を受信するための WebSocket エンドポイント
// @Tags websocket
// @Param token query string true "JWT アクセストークン"
// @Success 101 {string} string "Switching Protocols"
// @Failure 401 {object} response.ErrorResponse
// @Router /ws/jobs [get]
func (h *WebSocketHandler) HandleJobs(c *gin.Context) {
	log := logger.FromContext(c.Request.Context())

	// クエリパラメータからトークンを取得
	token := c.Query("token")
	if token == "" {
		Error(c, apperror.ErrUnauthorized.WithMessage("token パラメータは必須です"))
		return
	}

	// JWT を検証
	claims, err := h.tokenManager.Validate(token)
	if err != nil {
		log.Warn("WebSocket トークンが無効です", "error", err)
		Error(c, apperror.ErrUnauthorized.WithMessage("無効なトークンです"))
		return
	}

	// WebSocket にアップグレード
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Error("WebSocket へのアップグレードに失敗しました", "error", err)
		return
	}

	log.Info("WebSocket クライアントが接続しました", "user_id", claims.UserID)

	// クライアントを登録して読み書きループを開始
	client := h.hub.RegisterClient(conn, claims.UserID)
	go client.WritePump()
	go client.ReadPump()
}
