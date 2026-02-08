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
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
	"github.com/siropaca/anycast-backend/internal/service"
)

// アクセストークンの有効期限
const accessTokenExpiration = 1 * time.Hour

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
		Error(c, apperror.ErrValidation.WithMessage(formatValidationError(err)))
		return
	}

	result, err := h.authService.Register(c.Request.Context(), req)
	if err != nil {
		Error(c, err)
		return
	}

	accessToken, err := h.tokenManager.Generate(result.User.ID.String(), accessTokenExpiration)
	if err != nil {
		logger.FromContext(c.Request.Context()).Error("failed to generate token", "error", err, "user_id", result.User.ID)
		Error(c, apperror.ErrInternal.WithMessage("トークンの生成に失敗しました").WithError(err))
		return
	}

	Success(c, http.StatusCreated, response.AuthResponse{
		User:         result.User,
		AccessToken:  accessToken,
		RefreshToken: result.RefreshToken,
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
		Error(c, apperror.ErrValidation.WithMessage(formatValidationError(err)))
		return
	}

	result, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		Error(c, err)
		return
	}

	accessToken, err := h.tokenManager.Generate(result.User.ID.String(), accessTokenExpiration)
	if err != nil {
		logger.FromContext(c.Request.Context()).Error("failed to generate token", "error", err, "user_id", result.User.ID)
		Error(c, apperror.ErrInternal.WithMessage("トークンの生成に失敗しました").WithError(err))
		return
	}

	Success(c, http.StatusOK, response.AuthResponse{
		User:         result.User,
		AccessToken:  accessToken,
		RefreshToken: result.RefreshToken,
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
		Error(c, apperror.ErrValidation.WithMessage(formatValidationError(err)))
		return
	}

	result, err := h.authService.OAuthGoogle(c.Request.Context(), req)
	if err != nil {
		Error(c, err)
		return
	}

	accessToken, err := h.tokenManager.Generate(result.User.ID.String(), accessTokenExpiration)
	if err != nil {
		logger.FromContext(c.Request.Context()).Error("failed to generate token", "error", err, "user_id", result.User.ID)
		Error(c, apperror.ErrInternal.WithMessage("トークンの生成に失敗しました").WithError(err))
		return
	}

	authResponse := response.AuthResponse{
		User:         result.User,
		AccessToken:  accessToken,
		RefreshToken: result.RefreshToken,
	}

	if result.IsCreated {
		Success(c, http.StatusCreated, authResponse)
	} else {
		Success(c, http.StatusOK, authResponse)
	}
}

// RefreshToken godoc
// @Summary トークンリフレッシュ
// @Description リフレッシュトークンを使って新しいアクセストークンとリフレッシュトークンを発行します
// @Tags auth
// @Accept json
// @Produce json
// @Param request body request.RefreshTokenRequest true "リフレッシュトークン"
// @Success 200 {object} response.TokenRefreshDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req request.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(formatValidationError(err)))
		return
	}

	result, err := h.authService.RefreshToken(c.Request.Context(), req)
	if err != nil {
		Error(c, err)
		return
	}

	accessToken, err := h.tokenManager.Generate(result.UserID.String(), accessTokenExpiration)
	if err != nil {
		logger.FromContext(c.Request.Context()).Error("failed to generate token", "error", err, "user_id", result.UserID)
		Error(c, apperror.ErrInternal.WithMessage("トークンの生成に失敗しました").WithError(err))
		return
	}

	Success(c, http.StatusOK, response.TokenRefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: result.RefreshToken,
	})
}

// Logout godoc
// @Summary ログアウト
// @Description リフレッシュトークンを無効化してログアウトします
// @Tags auth
// @Accept json
// @Produce json
// @Param request body request.LogoutRequest true "ログアウトリクエスト"
// @Security BearerAuth
// @Success 204
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	var req request.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(formatValidationError(err)))
		return
	}

	if err := h.authService.Logout(c.Request.Context(), userID, req); err != nil {
		Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// ChangePassword godoc
// @Summary パスワード更新
// @Description 認証済みユーザーのパスワードを更新します
// @Tags auth
// @Accept json
// @Produce json
// @Param request body request.ChangePasswordRequest true "パスワード更新リクエスト"
// @Security BearerAuth
// @Success 204
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /auth/password [put]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	var req request.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(formatValidationError(err)))
		return
	}

	if err := h.authService.ChangePassword(c.Request.Context(), userID, req); err != nil {
		Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
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

// UpdateMe godoc
// @Summary ユーザー情報更新
// @Description ユーザーのプロフィール情報（表示名、自己紹介、アバター画像、ヘッダー画像）を更新します
// @Tags me
// @Accept json
// @Produce json
// @Param request body request.UpdateMeRequest true "ユーザー情報更新リクエスト"
// @Security BearerAuth
// @Success 200 {object} response.MeDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /me [patch]
func (h *AuthHandler) UpdateMe(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	var req request.UpdateMeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(formatValidationError(err)))
		return
	}

	me, err := h.authService.UpdateMe(c.Request.Context(), userID, req)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, http.StatusOK, me)
}

// DeleteMe godoc
// @Summary アカウント削除
// @Description 認証済みユーザーのアカウントを削除します
// @Tags me
// @Security BearerAuth
// @Success 204
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /me [delete]
func (h *AuthHandler) DeleteMe(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	if err := h.authService.DeleteMe(c.Request.Context(), userID); err != nil {
		Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// UpdateUsername godoc
// @Summary ユーザー名変更
// @Description ユーザー名を変更します
// @Tags me
// @Accept json
// @Produce json
// @Param request body request.UpdateUsernameRequest true "ユーザー名変更リクエスト"
// @Security BearerAuth
// @Success 200 {object} response.MeDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /me/username [patch]
func (h *AuthHandler) UpdateUsername(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	var req request.UpdateUsernameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(formatValidationError(err)))
		return
	}

	me, err := h.authService.UpdateUsername(c.Request.Context(), userID, req)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, http.StatusOK, me)
}

// CheckUsernameAvailability godoc
// @Summary ユーザー名利用可否チェック
// @Description 指定したユーザー名が利用可能かどうかを確認します
// @Tags me
// @Produce json
// @Param username query string true "チェック対象のユーザー名"
// @Security BearerAuth
// @Success 200 {object} response.UsernameCheckDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /me/username/check [get]
func (h *AuthHandler) CheckUsernameAvailability(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	var req request.CheckUsernameRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(formatValidationError(err)))
		return
	}

	result, err := h.authService.CheckUsernameAvailability(c.Request.Context(), userID, req)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, http.StatusOK, result)
}

// UpdatePrompt godoc
// @Summary ユーザープロンプト更新
// @Description ユーザーの台本生成用プロンプト（基本方針）を更新します
// @Tags me
// @Accept json
// @Produce json
// @Param request body request.UpdateUserPromptRequest true "プロンプト更新リクエスト"
// @Security BearerAuth
// @Success 200 {object} response.MeDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /me/prompt [patch]
func (h *AuthHandler) UpdatePrompt(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	var req request.UpdateUserPromptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(formatValidationError(err)))
		return
	}

	me, err := h.authService.UpdatePrompt(c.Request.Context(), userID, req)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, http.StatusOK, me)
}
