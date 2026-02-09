package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/repository"
	"github.com/siropaca/anycast-backend/internal/service"
)

// 検索関連のハンドラー
type SearchHandler struct {
	searchService service.SearchService
}

// SearchHandler を作成する
func NewSearchHandler(ss service.SearchService) *SearchHandler {
	return &SearchHandler{searchService: ss}
}

// SearchChannels godoc
// @Summary チャンネル検索
// @Description 公開中のチャンネルをキーワードで検索します。name, description を対象にフリーワード検索を行います。
// @Tags search
// @Accept json
// @Produce json
// @Param q query string true "検索キーワード"
// @Param categorySlug query string false "カテゴリスラッグでフィルタ"
// @Param limit query int false "取得件数（デフォルト: 20、最大: 100）"
// @Param offset query int false "オフセット（デフォルト: 0）"
// @Success 200 {object} response.SearchChannelListResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /search/channels [get]
func (h *SearchHandler) SearchChannels(c *gin.Context) {
	var req request.SearchChannelsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(formatValidationError(err)))
		return
	}

	filter := repository.SearchChannelFilter{
		Query:        req.Q,
		CategorySlug: req.CategorySlug,
		Limit:        req.Limit,
		Offset:       req.Offset,
	}

	result, err := h.searchService.SearchChannels(c.Request.Context(), filter)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}
