package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/middleware"
	"github.com/siropaca/anycast-backend/internal/service"
)

// 台本関連のハンドラー
type ScriptHandler struct {
	scriptService service.ScriptService
}

// ScriptHandler を作成する
func NewScriptHandler(ss service.ScriptService) *ScriptHandler {
	return &ScriptHandler{scriptService: ss}
}

// GenerateScript godoc
// @Summary 台本を AI で生成
// @Description 指定したエピソードの台本を AI を使って生成します。既存の台本がある場合は全て削除されます。
// @Tags script
// @Accept json
// @Produce json
// @Param channelId path string true "チャンネル ID"
// @Param episodeId path string true "エピソード ID"
// @Param request body request.GenerateScriptRequest true "台本生成リクエスト"
// @Success 200 {object} response.ScriptLineListResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /channels/{channelId}/episodes/{episodeId}/script/generate [post]
func (h *ScriptHandler) GenerateScript(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	channelID := c.Param("channelId")
	if channelID == "" {
		Error(c, apperror.ErrValidation.WithMessage("channelId は必須です"))
		return
	}

	episodeID := c.Param("episodeId")
	if episodeID == "" {
		Error(c, apperror.ErrValidation.WithMessage("episodeId は必須です"))
		return
	}

	var req request.GenerateScriptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(err.Error()))
		return
	}

	result, err := h.scriptService.GenerateScript(
		c.Request.Context(),
		userID,
		channelID,
		episodeID,
		req.Prompt,
		req.DurationMinutes,
		req.WithEmotion,
	)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// ImportScript godoc
// @Summary 台本テキストを取り込み
// @Description テキスト形式の台本をインポートします。既存の台本がある場合は全て削除されます。
// @Tags script
// @Accept json
// @Produce json
// @Param channelId path string true "チャンネル ID"
// @Param episodeId path string true "エピソード ID"
// @Param request body request.ImportScriptRequest true "台本インポートリクエスト"
// @Success 200 {object} response.ScriptLineListResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /channels/{channelId}/episodes/{episodeId}/script/import [post]
func (h *ScriptHandler) ImportScript(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	channelID := c.Param("channelId")
	if channelID == "" {
		Error(c, apperror.ErrValidation.WithMessage("channelId は必須です"))
		return
	}

	episodeID := c.Param("episodeId")
	if episodeID == "" {
		Error(c, apperror.ErrValidation.WithMessage("episodeId は必須です"))
		return
	}

	var req request.ImportScriptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(err.Error()))
		return
	}

	result, err := h.scriptService.ImportScript(
		c.Request.Context(),
		userID,
		channelID,
		episodeID,
		req.Text,
	)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// ExportScript godoc
// @Summary 台本テキストを出力
// @Description 台本をテキストファイルとしてダウンロードします。
// @Tags script
// @Produce text/plain
// @Param channelId path string true "チャンネル ID"
// @Param episodeId path string true "エピソード ID"
// @Success 200 {string} string "台本テキスト"
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /channels/{channelId}/episodes/{episodeId}/script/export [get]
func (h *ScriptHandler) ExportScript(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	channelID := c.Param("channelId")
	if channelID == "" {
		Error(c, apperror.ErrValidation.WithMessage("channelId は必須です"))
		return
	}

	episodeID := c.Param("episodeId")
	if episodeID == "" {
		Error(c, apperror.ErrValidation.WithMessage("episodeId は必須です"))
		return
	}

	result, err := h.scriptService.ExportScript(
		c.Request.Context(),
		userID,
		channelID,
		episodeID,
	)
	if err != nil {
		Error(c, err)
		return
	}

	// ファイル名を生成（ファイル名に使えない文字を除去）
	filename := sanitizeFilename(result.EpisodeTitle) + ".txt"

	// ファイルダウンロードとして返す
	// RFC 5987 に準拠した filename* を使用して日本語ファイル名に対応
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q; filename*=UTF-8''%s",
		filename,
		url.PathEscape(filename),
	))
	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(result.Text))
}

// ファイル名に使えない文字を除去・置換する
func sanitizeFilename(name string) string {
	// ファイル名に使えない文字を置換
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	sanitized := replacer.Replace(name)

	// 空白をトリム
	sanitized = strings.TrimSpace(sanitized)

	// 空になった場合はデフォルト名
	if sanitized == "" {
		return "script"
	}

	return sanitized
}
