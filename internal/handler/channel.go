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

// チャンネル関連のハンドラー
type ChannelHandler struct {
	channelService service.ChannelService
}

// ChannelHandler を作成する
func NewChannelHandler(cs service.ChannelService) *ChannelHandler {
	return &ChannelHandler{channelService: cs}
}

// ListMyChannels godoc
// @Summary 自分のチャンネル一覧取得
// @Description 認証ユーザーの所有するチャンネル一覧を取得します（非公開含む）
// @Tags me
// @Accept json
// @Produce json
// @Param status query string false "公開状態でフィルタ（published / draft）"
// @Param limit query int false "取得件数（デフォルト: 20、最大: 100）"
// @Param offset query int false "オフセット（デフォルト: 0）"
// @Success 200 {object} response.ChannelListWithPaginationResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /me/channels [get]
func (h *ChannelHandler) ListMyChannels(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	var req request.ListMyChannelsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(formatValidationError(err)))
		return
	}

	filter := repository.ChannelFilter{
		Status: req.Status,
		Limit:  req.Limit,
		Offset: req.Offset,
	}

	result, err := h.channelService.ListMyChannels(c.Request.Context(), userID, filter)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetMyChannel godoc
// @Summary 自分のチャンネル取得
// @Description 認証ユーザーの所有するチャンネルを取得します（非公開含む）
// @Tags me
// @Accept json
// @Produce json
// @Param channelId path string true "チャンネル ID"
// @Success 200 {object} response.ChannelDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /me/channels/{channelId} [get]
func (h *ChannelHandler) GetMyChannel(c *gin.Context) {
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

	result, err := h.channelService.GetMyChannel(c.Request.Context(), userID, channelID)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// CreateChannel godoc
// @Summary チャンネル作成
// @Description 新しいチャンネルを作成します
// @Tags channels
// @Accept json
// @Produce json
// @Param request body request.CreateChannelRequest true "チャンネル作成リクエスト"
// @Success 201 {object} response.ChannelDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /channels [post]
func (h *ChannelHandler) CreateChannel(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	var req request.CreateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(formatValidationError(err)))
		return
	}

	result, err := h.channelService.CreateChannel(c.Request.Context(), userID, req)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusCreated, result)
}

// GetChannel godoc
// @Summary チャンネル取得
// @Description チャンネルを取得します。認証なしでは公開中のチャンネルのみ、認証ありでは自分のチャンネルも取得可能です。
// @Tags channels
// @Accept json
// @Produce json
// @Param channelId path string true "チャンネル ID"
// @Success 200 {object} response.ChannelDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /channels/{channelId} [get]
func (h *ChannelHandler) GetChannel(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	channelID := c.Param("channelId")
	if channelID == "" {
		Error(c, apperror.ErrValidation.WithMessage("channelId は必須です"))
		return
	}

	result, err := h.channelService.GetChannel(c.Request.Context(), userID, channelID)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// UpdateChannel godoc
// @Summary チャンネル更新
// @Description チャンネルを更新します（オーナーのみ）
// @Tags channels
// @Accept json
// @Produce json
// @Param channelId path string true "チャンネル ID"
// @Param request body request.UpdateChannelRequest true "チャンネル更新リクエスト"
// @Success 200 {object} response.ChannelDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /channels/{channelId} [patch]
func (h *ChannelHandler) UpdateChannel(c *gin.Context) {
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

	var req request.UpdateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(formatValidationError(err)))
		return
	}

	result, err := h.channelService.UpdateChannel(c.Request.Context(), userID, channelID, req)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// DeleteChannel godoc
// @Summary チャンネル削除
// @Description チャンネルを削除します（オーナーのみ）
// @Tags channels
// @Param channelId path string true "チャンネル ID"
// @Success 204 "No Content"
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /channels/{channelId} [delete]
func (h *ChannelHandler) DeleteChannel(c *gin.Context) {
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

	if err := h.channelService.DeleteChannel(c.Request.Context(), userID, channelID); err != nil {
		Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// PublishChannel godoc
// @Summary チャンネル公開
// @Description 指定したチャンネルを公開します。publishedAt を省略すると現在時刻で即時公開、指定すると予約公開になります。
// @Tags channels
// @Accept json
// @Produce json
// @Param channelId path string true "チャンネル ID"
// @Param request body request.PublishChannelRequest false "公開リクエスト"
// @Success 200 {object} response.ChannelDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /channels/{channelId}/publish [post]
func (h *ChannelHandler) PublishChannel(c *gin.Context) {
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

	var req request.PublishChannelRequest
	// ボディが空でもエラーにならないよう ShouldBindJSON を使用
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		Error(c, apperror.ErrValidation.WithMessage(formatValidationError(err)))
		return
	}

	result, err := h.channelService.PublishChannel(c.Request.Context(), userID, channelID, req.PublishedAt)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// UnpublishChannel godoc
// @Summary チャンネル非公開
// @Description 指定したチャンネルを非公開（下書き）状態に戻します
// @Tags channels
// @Accept json
// @Produce json
// @Param channelId path string true "チャンネル ID"
// @Success 200 {object} response.ChannelDataResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /channels/{channelId}/unpublish [post]
func (h *ChannelHandler) UnpublishChannel(c *gin.Context) {
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

	result, err := h.channelService.UnpublishChannel(c.Request.Context(), userID, channelID)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// SetUserPrompt godoc
// @Summary チャンネルの台本プロンプト設定
// @Description 指定したチャンネルに台本プロンプトを設定します（オーナーのみ）
// @Tags channels
// @Accept json
// @Produce json
// @Param channelId path string true "チャンネル ID"
// @Param request body request.SetUserPromptRequest true "台本プロンプト設定リクエスト"
// @Success 200 {object} response.ChannelDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /channels/{channelId}/user-prompt [put]
func (h *ChannelHandler) SetUserPrompt(c *gin.Context) {
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

	var req request.SetUserPromptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(formatValidationError(err)))
		return
	}

	result, err := h.channelService.SetUserPrompt(c.Request.Context(), userID, channelID, req)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// SetDefaultBgm godoc
// @Summary チャンネルのデフォルト BGM 設定
// @Description 指定したチャンネルにデフォルト BGM を設定します。ユーザー BGM またはシステム BGM のどちらかを指定します。
// @Tags channels
// @Accept json
// @Produce json
// @Param channelId path string true "チャンネル ID"
// @Param request body request.SetDefaultBgmRequest true "デフォルト BGM 設定リクエスト"
// @Success 200 {object} response.ChannelDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /channels/{channelId}/default-bgm [put]
func (h *ChannelHandler) SetDefaultBgm(c *gin.Context) {
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

	var req request.SetDefaultBgmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(formatValidationError(err)))
		return
	}

	result, err := h.channelService.SetDefaultBgm(c.Request.Context(), userID, channelID, req)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// DeleteDefaultBgm godoc
// @Summary チャンネルのデフォルト BGM 削除
// @Description 指定したチャンネルのデフォルト BGM 設定を削除します
// @Tags channels
// @Accept json
// @Produce json
// @Param channelId path string true "チャンネル ID"
// @Success 200 {object} response.ChannelDataResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /channels/{channelId}/default-bgm [delete]
func (h *ChannelHandler) DeleteDefaultBgm(c *gin.Context) {
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

	result, err := h.channelService.DeleteDefaultBgm(c.Request.Context(), userID, channelID)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}
