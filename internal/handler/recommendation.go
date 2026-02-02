package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/middleware"
	"github.com/siropaca/anycast-backend/internal/service"
)

// おすすめ関連のハンドラー
type RecommendationHandler struct {
	recommendationService service.RecommendationService
}

// RecommendationHandler を作成する
func NewRecommendationHandler(rs service.RecommendationService) *RecommendationHandler {
	return &RecommendationHandler{recommendationService: rs}
}

// GetRecommendedChannels godoc
// @Summary おすすめチャンネル一覧取得
// @Description おすすめチャンネル一覧を取得します。未ログイン時は人気順・新着順、ログイン時はパーソナライズされた結果を返します。
// @Tags recommendations
// @Accept json
// @Produce json
// @Param categorySlug query string false "カテゴリスラッグでフィルタ"
// @Param limit query int false "取得件数（デフォルト: 20、最大: 50）"
// @Param offset query int false "オフセット（デフォルト: 0）"
// @Success 200 {object} response.RecommendedChannelListResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /recommendations/channels [get]
func (h *RecommendationHandler) GetRecommendedChannels(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req request.RecommendChannelsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(formatValidationError(err)))
		return
	}

	var userIDPtr *string
	if userID != "" {
		userIDPtr = &userID
	}

	result, err := h.recommendationService.GetRecommendedChannels(c.Request.Context(), userIDPtr, req)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetRecommendedEpisodes godoc
// @Summary おすすめエピソード一覧取得
// @Description おすすめエピソード一覧を取得します。未ログイン時は人気順・新着順、ログイン時は途中再生・デフォルトプレイリスト（後で聴く）・パーソナライズに基づく結果を返します。
// @Tags recommendations
// @Accept json
// @Produce json
// @Param categorySlug query string false "カテゴリスラッグでフィルタ（チャンネルのカテゴリ）"
// @Param limit query int false "取得件数（デフォルト: 20、最大: 50）"
// @Param offset query int false "オフセット（デフォルト: 0）"
// @Success 200 {object} response.RecommendedEpisodeListResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /recommendations/episodes [get]
func (h *RecommendationHandler) GetRecommendedEpisodes(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req request.RecommendEpisodesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(formatValidationError(err)))
		return
	}

	var userIDPtr *string
	if userID != "" {
		userIDPtr = &userID
	}

	result, err := h.recommendationService.GetRecommendedEpisodes(c.Request.Context(), userIDPtr, req)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}
