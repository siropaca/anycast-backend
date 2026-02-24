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
		Error(c, apperror.ErrValidation.WithMessage("channelId は必須です"))
		return
	}

	episodeID := c.Param("episodeId")
	if episodeID == "" {
		Error(c, apperror.ErrValidation.WithMessage("episodeId は必須です"))
		return
	}

	result, err := h.scriptLineService.ListByEpisodeID(c.Request.Context(), userID, channelID, episodeID)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// CreateScriptLine godoc
// @Summary 台本行作成
// @Description 指定した位置に新しい台本行を追加します
// @Tags script
// @Accept json
// @Produce json
// @Param channelId path string true "チャンネル ID"
// @Param episodeId path string true "エピソード ID"
// @Param request body request.CreateScriptLineRequest true "台本行作成リクエスト"
// @Success 201 {object} response.ScriptLineResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /channels/{channelId}/episodes/{episodeId}/script/lines [post]
func (h *ScriptLineHandler) CreateScriptLine(c *gin.Context) {
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

	var req request.CreateScriptLineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(formatValidationError(err)))
		return
	}

	result, err := h.scriptLineService.Create(c.Request.Context(), userID, channelID, episodeID, req)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusCreated, result)
}

// UpdateScriptLine godoc
// @Summary 台本行更新
// @Description 指定した台本行を更新します。
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
		Error(c, apperror.ErrValidation.WithMessage("channelId は必須です"))
		return
	}

	episodeID := c.Param("episodeId")
	if episodeID == "" {
		Error(c, apperror.ErrValidation.WithMessage("episodeId は必須です"))
		return
	}

	lineID := c.Param("lineId")
	if lineID == "" {
		Error(c, apperror.ErrValidation.WithMessage("lineId は必須です"))
		return
	}

	var req request.UpdateScriptLineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(formatValidationError(err)))
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
		Error(c, apperror.ErrValidation.WithMessage("channelId は必須です"))
		return
	}

	episodeID := c.Param("episodeId")
	if episodeID == "" {
		Error(c, apperror.ErrValidation.WithMessage("episodeId は必須です"))
		return
	}

	lineID := c.Param("lineId")
	if lineID == "" {
		Error(c, apperror.ErrValidation.WithMessage("lineId は必須です"))
		return
	}

	if err := h.scriptLineService.Delete(c.Request.Context(), userID, channelID, episodeID, lineID); err != nil {
		Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// DeleteAllScriptLines godoc
// @Summary 台本行全削除
// @Description 指定したエピソードの台本行をすべて削除します
// @Tags script
// @Accept json
// @Produce json
// @Param channelId path string true "チャンネル ID"
// @Param episodeId path string true "エピソード ID"
// @Success 204 "No Content"
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /channels/{channelId}/episodes/{episodeId}/script/lines [delete]
func (h *ScriptLineHandler) DeleteAllScriptLines(c *gin.Context) {
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

	if err := h.scriptLineService.DeleteAll(c.Request.Context(), userID, channelID, episodeID); err != nil {
		Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// ReorderScriptLines godoc
// @Summary 台本行並び替え
// @Description 台本行の順序を並び替えます
// @Tags script
// @Accept json
// @Produce json
// @Param channelId path string true "チャンネル ID"
// @Param episodeId path string true "エピソード ID"
// @Param request body request.ReorderScriptLinesRequest true "並び替えリクエスト"
// @Success 200 {object} response.ScriptLineListResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /channels/{channelId}/episodes/{episodeId}/script/reorder [post]
func (h *ScriptLineHandler) ReorderScriptLines(c *gin.Context) {
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

	var req request.ReorderScriptLinesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(formatValidationError(err)))
		return
	}

	result, err := h.scriptLineService.Reorder(c.Request.Context(), userID, channelID, episodeID, req)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}
