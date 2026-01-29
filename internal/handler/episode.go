package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/middleware"
	"github.com/siropaca/anycast-backend/internal/repository"
	"github.com/siropaca/anycast-backend/internal/service"
)

// エピソード関連のハンドラー
type EpisodeHandler struct {
	episodeService service.EpisodeService
}

// EpisodeHandler を作成する
func NewEpisodeHandler(es service.EpisodeService) *EpisodeHandler {
	return &EpisodeHandler{episodeService: es}
}

// ListMyChannelEpisodes godoc
// @Summary 自分のチャンネルのエピソード一覧取得
// @Description 自分のチャンネルに紐付くエピソード一覧を取得します（非公開含む）
// @Tags me
// @Accept json
// @Produce json
// @Param channelId path string true "チャンネル ID"
// @Param status query string false "公開状態でフィルタ（published / draft）"
// @Param limit query int false "取得件数（デフォルト: 20、最大: 100）"
// @Param offset query int false "オフセット（デフォルト: 0）"
// @Success 200 {object} response.EpisodeListWithPaginationResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /me/channels/{channelId}/episodes [get]
func (h *EpisodeHandler) ListMyChannelEpisodes(c *gin.Context) {
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

	var req request.ListMyChannelEpisodesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(formatValidationError(err)))
		return
	}

	filter := repository.EpisodeFilter{
		Status: req.Status,
		Limit:  req.Limit,
		Offset: req.Offset,
	}

	result, err := h.episodeService.ListMyChannelEpisodes(c.Request.Context(), userID, channelID, filter)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetMyChannelEpisode godoc
// @Summary 自分のチャンネルのエピソード取得
// @Description 自分のチャンネルに紐付くエピソードを取得します（非公開含む）
// @Tags me
// @Accept json
// @Produce json
// @Param channelId path string true "チャンネル ID"
// @Param episodeId path string true "エピソード ID"
// @Success 200 {object} response.EpisodeDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /me/channels/{channelId}/episodes/{episodeId} [get]
func (h *EpisodeHandler) GetMyChannelEpisode(c *gin.Context) {
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

	result, err := h.episodeService.GetMyChannelEpisode(c.Request.Context(), userID, channelID, episodeID)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// CreateEpisode godoc
// @Summary エピソード作成
// @Description 指定したチャンネルにエピソードを作成します
// @Tags episodes
// @Accept json
// @Produce json
// @Param channelId path string true "チャンネル ID"
// @Param request body request.CreateEpisodeRequest true "エピソード作成リクエスト"
// @Success 201 {object} response.EpisodeDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /channels/{channelId}/episodes [post]
func (h *EpisodeHandler) CreateEpisode(c *gin.Context) {
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

	var req request.CreateEpisodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(formatValidationError(err)))
		return
	}

	result, err := h.episodeService.CreateEpisode(
		c.Request.Context(),
		userID,
		channelID,
		req.Title,
		req.Description,
		req.ArtworkImageID,
	)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusCreated, map[string]any{"data": result})
}

// UpdateEpisode godoc
// @Summary エピソード更新
// @Description 指定したエピソードを更新します
// @Tags episodes
// @Accept json
// @Produce json
// @Param channelId path string true "チャンネル ID"
// @Param episodeId path string true "エピソード ID"
// @Param request body request.UpdateEpisodeRequest true "エピソード更新リクエスト"
// @Success 200 {object} response.EpisodeDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /channels/{channelId}/episodes/{episodeId} [patch]
func (h *EpisodeHandler) UpdateEpisode(c *gin.Context) {
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

	var req request.UpdateEpisodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(formatValidationError(err)))
		return
	}

	result, err := h.episodeService.UpdateEpisode(c.Request.Context(), userID, channelID, episodeID, req)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// DeleteEpisode godoc
// @Summary エピソード削除
// @Description 指定したエピソードを削除します
// @Tags episodes
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
// @Router /channels/{channelId}/episodes/{episodeId} [delete]
func (h *EpisodeHandler) DeleteEpisode(c *gin.Context) {
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

	if err := h.episodeService.DeleteEpisode(c.Request.Context(), userID, channelID, episodeID); err != nil {
		Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// PublishEpisode godoc
// @Summary エピソード公開
// @Description 指定したエピソードを公開します。publishedAt を省略すると現在時刻で即時公開、指定すると予約公開になります。
// @Tags episodes
// @Accept json
// @Produce json
// @Param channelId path string true "チャンネル ID"
// @Param episodeId path string true "エピソード ID"
// @Param request body request.PublishEpisodeRequest false "公開リクエスト"
// @Success 200 {object} response.EpisodeDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /channels/{channelId}/episodes/{episodeId}/publish [post]
func (h *EpisodeHandler) PublishEpisode(c *gin.Context) {
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

	var req request.PublishEpisodeRequest
	// ボディが空でもエラーにならないよう ShouldBindJSON を使用
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		Error(c, apperror.ErrValidation.WithMessage(formatValidationError(err)))
		return
	}

	result, err := h.episodeService.PublishEpisode(c.Request.Context(), userID, channelID, episodeID, req.PublishedAt)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// UnpublishEpisode godoc
// @Summary エピソード非公開
// @Description 指定したエピソードを非公開（下書き）状態に戻します
// @Tags episodes
// @Accept json
// @Produce json
// @Param channelId path string true "チャンネル ID"
// @Param episodeId path string true "エピソード ID"
// @Success 200 {object} response.EpisodeDataResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /channels/{channelId}/episodes/{episodeId}/unpublish [post]
func (h *EpisodeHandler) UnpublishEpisode(c *gin.Context) {
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

	result, err := h.episodeService.UnpublishEpisode(c.Request.Context(), userID, channelID, episodeID)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// SetEpisodeBgm godoc
// @Summary エピソード BGM 設定
// @Description 指定したエピソードに BGM を設定します。ユーザー BGM またはシステム BGM のどちらかを指定します。
// @Tags episodes
// @Accept json
// @Produce json
// @Param channelId path string true "チャンネル ID"
// @Param episodeId path string true "エピソード ID"
// @Param request body request.SetEpisodeBgmRequest true "BGM 設定リクエスト"
// @Success 200 {object} response.EpisodeDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /channels/{channelId}/episodes/{episodeId}/bgm [put]
func (h *EpisodeHandler) SetEpisodeBgm(c *gin.Context) {
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

	var req request.SetEpisodeBgmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(formatValidationError(err)))
		return
	}

	result, err := h.episodeService.SetEpisodeBgm(c.Request.Context(), userID, channelID, episodeID, req)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// IncrementPlayCount godoc
// @Summary 再生回数カウント
// @Description エピソードの再生回数をインクリメントします。クライアントは再生開始から30秒経過した時点で呼び出します。
// @Tags episodes
// @Accept json
// @Produce json
// @Param episodeId path string true "エピソード ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /episodes/{episodeId}/play [post]
func (h *EpisodeHandler) IncrementPlayCount(c *gin.Context) {
	episodeID := c.Param("episodeId")
	if episodeID == "" {
		Error(c, apperror.ErrValidation.WithMessage("episodeId は必須です"))
		return
	}

	if err := h.episodeService.IncrementPlayCount(c.Request.Context(), episodeID); err != nil {
		Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// DeleteEpisodeBgm godoc
// @Summary エピソード BGM 削除
// @Description 指定したエピソードに設定されている BGM を削除します
// @Tags episodes
// @Accept json
// @Produce json
// @Param channelId path string true "チャンネル ID"
// @Param episodeId path string true "エピソード ID"
// @Success 200 {object} response.EpisodeDataResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /channels/{channelId}/episodes/{episodeId}/bgm [delete]
func (h *EpisodeHandler) DeleteEpisodeBgm(c *gin.Context) {
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

	result, err := h.episodeService.DeleteEpisodeBgm(c.Request.Context(), userID, channelID, episodeID)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}
