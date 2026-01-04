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
		Error(c, apperror.ErrValidation.WithMessage(err.Error()))
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
		Error(c, apperror.ErrValidation.WithMessage("channelId is required"))
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
		Error(c, apperror.ErrValidation.WithMessage(err.Error()))
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
// @Description チャンネルを取得します（公開中、または自分のチャンネルのみ）
// @Tags channels
// @Accept json
// @Produce json
// @Param channelId path string true "チャンネル ID"
// @Success 200 {object} response.ChannelDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /channels/{channelId} [get]
func (h *ChannelHandler) GetChannel(c *gin.Context) {
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
		Error(c, apperror.ErrValidation.WithMessage("channelId is required"))
		return
	}

	var req request.UpdateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(err.Error()))
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
		Error(c, apperror.ErrValidation.WithMessage("channelId is required"))
		return
	}

	if err := h.channelService.DeleteChannel(c.Request.Context(), userID, channelID); err != nil {
		Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
