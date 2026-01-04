package handler

import (
	"net/http"

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
		Error(c, apperror.ErrValidation.WithMessage("channelId is required"))
		return
	}

	episodeID := c.Param("episodeId")
	if episodeID == "" {
		Error(c, apperror.ErrValidation.WithMessage("episodeId is required"))
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
	)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}
