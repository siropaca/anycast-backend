package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/middleware"
	"github.com/siropaca/anycast-backend/internal/service"
)

// 台本行関連のハンドラー
type ScriptLineHandler struct {
	scriptLineService service.ScriptLineService
}

// ScriptLineHandler を作成する
func NewScriptLineHandler(sls service.ScriptLineService) *ScriptLineHandler {
	return &ScriptLineHandler{scriptLineService: sls}
}

// ListScriptLines godoc
// @Summary 台本行一覧取得
// @Description 指定したエピソードの台本行一覧を取得します
// @Tags script
// @Accept json
// @Produce json
// @Param channelId path string true "チャンネル ID"
// @Param episodeId path string true "エピソード ID"
// @Success 200 {object} response.ScriptLineListResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /channels/{channelId}/episodes/{episodeId}/script/lines [get]
func (h *ScriptLineHandler) ListScriptLines(c *gin.Context) {
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

	result, err := h.scriptLineService.ListByEpisodeID(c.Request.Context(), userID, channelID, episodeID)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}
