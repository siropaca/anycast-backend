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
// @Param username path string true "ユーザー名"
// @Success 200 {object} response.PublicUserDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /users/{username} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		Error(c, apperror.ErrValidation.WithMessage("username は必須です"))
		return
	}

	result, err := h.userService.GetUser(c.Request.Context(), username)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}
