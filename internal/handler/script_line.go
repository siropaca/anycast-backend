package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
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

// UpdateScriptLine godoc
// @Summary 台本行更新
// @Description 指定した台本行を更新します。speech 行のみ対応しています。
// @Tags script
// @Accept json
// @Produce json
// @Param channelId path string true "チャンネル ID"
// @Param episodeId path string true "エピソード ID"
// @Param lineId path string true "台本行 ID"
// @Param request body request.UpdateScriptLineRequest true "台本行更新リクエスト"
// @Success 200 {object} response.ScriptLineResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /channels/{channelId}/episodes/{episodeId}/script/lines/{lineId} [patch]
func (h *ScriptLineHandler) UpdateScriptLine(c *gin.Context) {
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

	lineID := c.Param("lineId")
	if lineID == "" {
		Error(c, apperror.ErrValidation.WithMessage("lineId is required"))
		return
	}

	var req request.UpdateScriptLineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(err.Error()))
		return
	}

	result, err := h.scriptLineService.Update(c.Request.Context(), userID, channelID, episodeID, lineID, req)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// DeleteScriptLine godoc
// @Summary 台本行削除
// @Description 指定した台本行を削除します
// @Tags script
// @Accept json
// @Produce json
// @Param channelId path string true "チャンネル ID"
// @Param episodeId path string true "エピソード ID"
// @Param lineId path string true "台本行 ID"
// @Success 204 "No Content"
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /channels/{channelId}/episodes/{episodeId}/script/lines/{lineId} [delete]
func (h *ScriptLineHandler) DeleteScriptLine(c *gin.Context) {
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

	lineID := c.Param("lineId")
	if lineID == "" {
		Error(c, apperror.ErrValidation.WithMessage("lineId is required"))
		return
	}

	if err := h.scriptLineService.Delete(c.Request.Context(), userID, channelID, episodeID, lineID); err != nil {
		Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// GenerateAudio godoc
// @Summary 台本行の音声を生成
// @Description 指定した台本行の音声を TTS で生成します。speech 行のみ対応しています。
// @Tags script
// @Accept json
// @Produce json
// @Param channelId path string true "チャンネル ID"
// @Param episodeId path string true "エピソード ID"
// @Param lineId path string true "台本行 ID"
// @Success 200 {object} response.GenerateAudioResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /channels/{channelId}/episodes/{episodeId}/script/lines/{lineId}/audio/generate [post]
func (h *ScriptLineHandler) GenerateAudio(c *gin.Context) {
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

	lineID := c.Param("lineId")
	if lineID == "" {
		Error(c, apperror.ErrValidation.WithMessage("lineId is required"))
		return
	}

	result, err := h.scriptLineService.GenerateAudio(c.Request.Context(), userID, channelID, episodeID, lineID)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}
