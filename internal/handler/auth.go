package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/middleware"
	"github.com/siropaca/anycast-backend/internal/pkg/jwt"
	"github.com/siropaca/anycast-backend/internal/service"
)

// トークンの有効期限
const tokenExpiration = 24 * time.Hour

// 認証関連のハンドラー
type AuthHandler struct {
	authService  service.AuthService
	tokenManager jwt.TokenManager
}

// AuthHandler を作成する
func NewAuthHandler(as service.AuthService, tm jwt.TokenManager) *AuthHandler {
	return &AuthHandler{
		authService:  as,
		tokenManager: tm,
	}
}

// Register godoc
// @Summary ユーザー登録
// @Description 新規ユーザーを登録します
// @Tags auth
// @Accept json
// @Produce json
// @Param request body request.RegisterRequest true "登録情報"
// @Success 201 {object} response.AuthDataResponse
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

	token, err := h.tokenManager.Generate(user.ID.String(), tokenExpiration)
	if err != nil {
		Error(c, apperror.ErrInternal.WithMessage("Failed to generate token").WithError(err))
		return
	}

	Success(c, http.StatusCreated, response.AuthResponse{
		User:  *user,
		Token: token,
	})
}

// Login godoc
// @Summary メール/パスワード認証
// @Description メールアドレスとパスワードで認証します
// @Tags auth
// @Accept json
// @Produce json
// @Param request body request.LoginRequest true "認証情報"
// @Success 200 {object} response.AuthDataResponse
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

	token, err := h.tokenManager.Generate(user.ID.String(), tokenExpiration)
	if err != nil {
		Error(c, apperror.ErrInternal.WithMessage("Failed to generate token").WithError(err))
		return
	}

	Success(c, http.StatusOK, response.AuthResponse{
		User:  *user,
		Token: token,
	})
}

// OAuthGoogle godoc
// @Summary Google OAuth 認証
// @Description Google OAuth でユーザーを認証/作成します
// @Tags auth
// @Accept json
// @Produce json
// @Param request body request.OAuthGoogleRequest true "OAuth 情報"
// @Success 200 {object} response.AuthDataResponse "既存ユーザー"
// @Success 201 {object} response.AuthDataResponse "新規ユーザー"
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

	token, err := h.tokenManager.Generate(result.User.ID.String(), tokenExpiration)
	if err != nil {
		Error(c, apperror.ErrInternal.WithMessage("Failed to generate token").WithError(err))
		return
	}

	authResponse := response.AuthResponse{
		User:  result.User,
		Token: token,
	}

	if result.IsCreated {
		Success(c, http.StatusCreated, authResponse)
	} else {
		Success(c, http.StatusOK, authResponse)
	}
}

// GetMe godoc
// @Summary 現在のユーザー取得
// @Description 認証済みユーザーの情報を取得します
// @Tags me
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.MeDataResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /me [get]
func (h *AuthHandler) GetMe(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	me, err := h.authService.GetMe(c.Request.Context(), userID)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, http.StatusOK, me)
}
