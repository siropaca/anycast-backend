package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/service"
)

// ユーザー関連のハンドラー
type UserHandler struct {
	userService service.UserService
}

// UserHandler を作成する
func NewUserHandler(us service.UserService) *UserHandler {
	return &UserHandler{userService: us}
}

// GetUser godoc
// @Summary ユーザー取得
// @Description 指定されたユーザーの公開プロフィールと公開チャンネル一覧を取得します
// @Tags users
// @Accept json
// @Produce json
// @Param userId path string true "ユーザー ID"
// @Success 200 {object} response.PublicUserDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /users/{userId} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		Error(c, apperror.ErrValidation.WithMessage("userId は必須です"))
		return
	}

	result, err := h.userService.GetUser(c.Request.Context(), userID)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}
