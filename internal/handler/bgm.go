package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/middleware"
	"github.com/siropaca/anycast-backend/internal/service"
)

// BGM 関連のハンドラー
type BgmHandler struct {
	bgmService service.BgmService
}

// BgmHandler を作成する
func NewBgmHandler(bs service.BgmService) *BgmHandler {
	return &BgmHandler{bgmService: bs}
}

// ListMyBgms godoc
// @Summary 自分の BGM 一覧取得
// @Description 認証ユーザーの所有する BGM 一覧を取得します。include_default=true の場合はデフォルト BGM も含めます。
// @Tags me
// @Accept json
// @Produce json
// @Param include_default query bool false "デフォルト BGM を含めるかどうか（デフォルト: false）"
// @Param limit query int false "取得件数（デフォルト: 20、最大: 100）"
// @Param offset query int false "オフセット（デフォルト: 0）"
// @Success 200 {object} response.BgmListWithPaginationResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /me/bgms [get]
func (h *BgmHandler) ListMyBgms(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	var req request.ListMyBgmsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(err.Error()))
		return
	}

	result, err := h.bgmService.ListMyBgms(c.Request.Context(), userID, req)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}
