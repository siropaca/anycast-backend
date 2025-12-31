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
