package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/middleware"
	"github.com/siropaca/anycast-backend/internal/service"
)

// フォロー関連のハンドラー
type FollowHandler struct {
	followService service.FollowService
}

// FollowHandler を作成する
func NewFollowHandler(fs service.FollowService) *FollowHandler {
	return &FollowHandler{followService: fs}
}

// ListFollows godoc
// @Summary フォロー中のユーザー一覧取得
// @Description 自分がフォロー中のユーザー一覧を取得します（フォロー日時の降順）
// @Tags me
// @Accept json
// @Produce json
// @Param limit query int false "取得件数（デフォルト: 20、最大: 100）"
// @Param offset query int false "オフセット（デフォルト: 0）"
// @Success 200 {object} response.FollowListWithPaginationResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /me/follows [get]
func (h *FollowHandler) ListFollows(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	var req request.ListFollowsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(formatValidationError(err)))
		return
	}

	result, err := h.followService.ListFollows(c.Request.Context(), userID, req.Limit, req.Offset)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// CreateFollow godoc
// @Summary フォロー登録
// @Description ユーザーをフォローします
// @Tags users
// @Accept json
// @Produce json
// @Param userId path string true "ユーザー ID"
// @Success 201 {object} response.FollowDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /users/{userId}/follow [post]
func (h *FollowHandler) CreateFollow(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	targetUserID := c.Param("userId")
	if targetUserID == "" {
		Error(c, apperror.ErrValidation.WithMessage("userId は必須です"))
		return
	}

	result, err := h.followService.CreateFollow(c.Request.Context(), userID, targetUserID)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusCreated, result)
}

// DeleteFollow godoc
// @Summary フォロー解除
// @Description ユーザーのフォローを解除します
// @Tags users
// @Accept json
// @Produce json
// @Param userId path string true "ユーザー ID"
// @Success 204 "No Content"
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /users/{userId}/follow [delete]
func (h *FollowHandler) DeleteFollow(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	targetUserID := c.Param("userId")
	if targetUserID == "" {
		Error(c, apperror.ErrValidation.WithMessage("userId は必須です"))
		return
	}

	if err := h.followService.DeleteFollow(c.Request.Context(), userID, targetUserID); err != nil {
		Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
