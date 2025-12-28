package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/service"
)

// 認証関連のハンドラー
type AuthHandler struct {
	authService service.AuthService
}

// AuthHandler を作成する
func NewAuthHandler(as service.AuthService) *AuthHandler {
	return &AuthHandler{authService: as}
}

// Register godoc
// @Summary ユーザー登録
// @Description 新規ユーザーを登録します
// @Tags auth
// @Accept json
// @Produce json
// @Param request body request.RegisterRequest true "登録情報"
// @Success 201 {object} response.UserDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req request.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(err.Error()))
		return
	}

	user, err := h.authService.Register(c.Request.Context(), req)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, http.StatusCreated, user)
}

// Login godoc
// @Summary メール/パスワード認証
// @Description メールアドレスとパスワードで認証します
// @Tags auth
// @Accept json
// @Produce json
// @Param request body request.LoginRequest true "認証情報"
// @Success 200 {object} response.UserDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req request.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(err.Error()))
		return
	}

	user, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, http.StatusOK, user)
}

// OAuthGoogle godoc
// @Summary Google OAuth 認証
// @Description Google OAuth でユーザーを認証/作成します
// @Tags auth
// @Accept json
// @Produce json
// @Param request body request.OAuthGoogleRequest true "OAuth 情報"
// @Success 200 {object} response.UserDataResponse "既存ユーザー"
// @Success 201 {object} response.UserDataResponse "新規ユーザー"
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /auth/oauth/google [post]
func (h *AuthHandler) OAuthGoogle(c *gin.Context) {
	var req request.OAuthGoogleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(err.Error()))
		return
	}

	result, err := h.authService.OAuthGoogle(c.Request.Context(), req)
	if err != nil {
		Error(c, err)
		return
	}

	if result.IsCreated {
		Success(c, http.StatusCreated, result.User)
	} else {
		Success(c, http.StatusOK, result.User)
	}
}
