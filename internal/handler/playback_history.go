package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/middleware"
	"github.com/siropaca/anycast-backend/internal/service"
)

// 再生履歴関連のハンドラー
type PlaybackHistoryHandler struct {
	playbackHistoryService service.PlaybackHistoryService
}

// PlaybackHistoryHandler を作成する
func NewPlaybackHistoryHandler(phs service.PlaybackHistoryService) *PlaybackHistoryHandler {
	return &PlaybackHistoryHandler{playbackHistoryService: phs}
}

// ListPlaybackHistory godoc
// @Summary 再生履歴一覧取得
// @Description 自分の再生履歴一覧を取得します（最近再生した順）
// @Tags me
// @Accept json
// @Produce json
// @Param completed query bool false "完了状態でフィルタ"
// @Param limit query int false "取得件数（デフォルト: 20、最大: 100）"
// @Param offset query int false "オフセット（デフォルト: 0）"
// @Success 200 {object} response.PlaybackHistoryListWithPaginationResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /me/playback-history [get]
func (h *PlaybackHistoryHandler) ListPlaybackHistory(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	var req request.ListPlaybackHistoryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(formatValidationError(err)))
		return
	}

	result, err := h.playbackHistoryService.ListPlaybackHistory(c.Request.Context(), userID, req.Completed, req.Limit, req.Offset)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// DeleteAllPlaybackHistory godoc
// @Summary 再生履歴をすべて削除
// @Description 認証済みユーザーの再生履歴をすべて削除します
// @Tags me
// @Security BearerAuth
// @Success 204
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /me/playback-history [delete]
func (h *PlaybackHistoryHandler) DeleteAllPlaybackHistory(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	if err := h.playbackHistoryService.DeleteAllPlaybackHistory(c.Request.Context(), userID); err != nil {
		Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// UpdatePlayback godoc
// @Summary 再生履歴を更新
// @Description 再生位置や完了状態を更新します（なければ作成）
// @Tags episodes
// @Accept json
// @Produce json
// @Param episodeId path string true "エピソード ID"
// @Param request body request.UpdatePlaybackRequest true "再生状態更新リクエスト"
// @Success 200 {object} response.PlaybackDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /episodes/{episodeId}/playback [put]
func (h *PlaybackHistoryHandler) UpdatePlayback(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	episodeID := c.Param("episodeId")
	if episodeID == "" {
		Error(c, apperror.ErrValidation.WithMessage("episodeId は必須です"))
		return
	}

	var req request.UpdatePlaybackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(formatValidationError(err)))
		return
	}

	result, err := h.playbackHistoryService.UpdatePlayback(c.Request.Context(), userID, episodeID, req)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// DeletePlayback godoc
// @Summary 再生履歴を削除
// @Description 指定したエピソードの再生履歴を削除します
// @Tags episodes
// @Accept json
// @Produce json
// @Param episodeId path string true "エピソード ID"
// @Success 204 "No Content"
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /episodes/{episodeId}/playback [delete]
func (h *PlaybackHistoryHandler) DeletePlayback(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	episodeID := c.Param("episodeId")
	if episodeID == "" {
		Error(c, apperror.ErrValidation.WithMessage("episodeId は必須です"))
		return
	}

	if err := h.playbackHistoryService.DeletePlayback(c.Request.Context(), userID, episodeID); err != nil {
		Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
