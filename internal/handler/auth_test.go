package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/middleware"
	"github.com/siropaca/anycast-backend/internal/pkg/jwt"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/service"
)

// AuthService のモック
type mockAuthService struct {
	mock.Mock
}

func (m *mockAuthService) Register(ctx context.Context, req request.RegisterRequest) (*service.AuthResult, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.AuthResult), args.Error(1)
}

func (m *mockAuthService) Login(ctx context.Context, req request.LoginRequest) (*service.AuthResult, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.AuthResult), args.Error(1)
}

func (m *mockAuthService) OAuthGoogle(ctx context.Context, req request.OAuthGoogleRequest) (*service.AuthResult, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.AuthResult), args.Error(1)
}

func (m *mockAuthService) RefreshToken(ctx context.Context, req request.RefreshTokenRequest) (*service.RefreshResult, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.RefreshResult), args.Error(1)
}

func (m *mockAuthService) Logout(ctx context.Context, userID string, req request.LogoutRequest) error {
	args := m.Called(ctx, userID, req)
	return args.Error(0)
}

func (m *mockAuthService) GetMe(ctx context.Context, userID string) (*response.MeResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.MeResponse), args.Error(1)
}

func (m *mockAuthService) UpdateMe(ctx context.Context, userID string, req request.UpdateMeRequest) (*response.MeResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.MeResponse), args.Error(1)
}

func (m *mockAuthService) UpdatePrompt(ctx context.Context, userID string, req request.UpdateUserPromptRequest) (*response.MeResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.MeResponse), args.Error(1)
}

// TokenManager のモック
type mockTokenManager struct {
	mock.Mock
}

func (m *mockTokenManager) Generate(userID string, expiration time.Duration) (string, error) {
	args := m.Called(userID, expiration)
	return args.String(0), args.Error(1)
}

func (m *mockTokenManager) Validate(tokenString string) (*jwt.Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwt.Claims), args.Error(1)
}

// テスト用のルーターをセットアップする
func setupAuthRouter(h *AuthHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/auth/register", h.Register)
	r.POST("/auth/login", h.Login)
	r.POST("/auth/oauth/google", h.OAuthGoogle)
	r.POST("/auth/refresh", h.RefreshToken)
	r.GET("/me", h.GetMe)
	return r
}

// 認証済みルーターをセットアップする
func setupAuthenticatedAuthRouter(h *AuthHandler, userID string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(string(middleware.UserIDKey), userID)
		c.Next()
	})
	r.GET("/me", h.GetMe)
	r.POST("/auth/logout", h.Logout)
	return r
}

// テスト用の AuthResult を生成する
func createTestAuthResult() *service.AuthResult {
	return &service.AuthResult{
		User: response.UserResponse{
			ID:          uuid.New(),
			Email:       "test@example.com",
			Username:    "testuser",
			DisplayName: "Test User",
			Role:        "user",
		},
		RefreshToken: "test-refresh-token",
	}
}

func TestAuthHandler_Register(t *testing.T) {
	t.Run("ユーザー登録が成功する", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		mockTM := new(mockTokenManager)

		result := createTestAuthResult()
		mockSvc.On("Register", mock.Anything, mock.AnythingOfType("request.RegisterRequest")).Return(result, nil)
		mockTM.On("Generate", result.User.ID.String(), accessTokenExpiration).Return("test-access-token", nil)

		handler := NewAuthHandler(mockSvc, mockTM)
		router := setupAuthRouter(handler)

		reqBody := request.RegisterRequest{
			Email:       "test@example.com",
			Password:    "password123",
			DisplayName: "Test User",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp map[string]response.AuthResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "test-access-token", resp["data"].AccessToken)
		assert.Equal(t, "test-refresh-token", resp["data"].RefreshToken)
		assert.Equal(t, result.User.Email, resp["data"].User.Email)
		mockSvc.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})

	t.Run("バリデーションエラーの場合は 400 を返す", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		mockTM := new(mockTokenManager)

		handler := NewAuthHandler(mockSvc, mockTM)
		router := setupAuthRouter(handler)

		// パスワードが短すぎる
		reqBody := map[string]string{
			"email":       "test@example.com",
			"password":    "short",
			"displayName": "Test",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("メールアドレスが既に使用されている場合は 409 を返す", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		mockTM := new(mockTokenManager)

		mockSvc.On("Register", mock.Anything, mock.AnythingOfType("request.RegisterRequest")).Return(nil, apperror.ErrDuplicateEmail)

		handler := NewAuthHandler(mockSvc, mockTM)
		router := setupAuthRouter(handler)

		reqBody := request.RegisterRequest{
			Email:       "existing@example.com",
			Password:    "password123",
			DisplayName: "Test User",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("トークン生成が失敗した場合は 500 を返す", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		mockTM := new(mockTokenManager)

		result := createTestAuthResult()
		mockSvc.On("Register", mock.Anything, mock.AnythingOfType("request.RegisterRequest")).Return(result, nil)
		mockTM.On("Generate", result.User.ID.String(), accessTokenExpiration).Return("", errors.New("token generation failed"))

		handler := NewAuthHandler(mockSvc, mockTM)
		router := setupAuthRouter(handler)

		reqBody := request.RegisterRequest{
			Email:       "test@example.com",
			Password:    "password123",
			DisplayName: "Test User",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})
}

func TestAuthHandler_Login(t *testing.T) {
	t.Run("ログインが成功する", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		mockTM := new(mockTokenManager)

		result := createTestAuthResult()
		mockSvc.On("Login", mock.Anything, mock.AnythingOfType("request.LoginRequest")).Return(result, nil)
		mockTM.On("Generate", result.User.ID.String(), accessTokenExpiration).Return("test-access-token", nil)

		handler := NewAuthHandler(mockSvc, mockTM)
		router := setupAuthRouter(handler)

		reqBody := request.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]response.AuthResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "test-access-token", resp["data"].AccessToken)
		assert.Equal(t, "test-refresh-token", resp["data"].RefreshToken)
		mockSvc.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})

	t.Run("バリデーションエラーの場合は 400 を返す", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		mockTM := new(mockTokenManager)

		handler := NewAuthHandler(mockSvc, mockTM)
		router := setupAuthRouter(handler)

		// メールアドレスが不正
		reqBody := map[string]string{
			"email":    "invalid-email",
			"password": "password123",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("認証が失敗した場合は 401 を返す", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		mockTM := new(mockTokenManager)

		mockSvc.On("Login", mock.Anything, mock.AnythingOfType("request.LoginRequest")).Return(nil, apperror.ErrUnauthorized.WithMessage("Invalid credentials"))

		handler := NewAuthHandler(mockSvc, mockTM)
		router := setupAuthRouter(handler)

		reqBody := request.LoginRequest{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("トークン生成が失敗した場合は 500 を返す", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		mockTM := new(mockTokenManager)

		result := createTestAuthResult()
		mockSvc.On("Login", mock.Anything, mock.AnythingOfType("request.LoginRequest")).Return(result, nil)
		mockTM.On("Generate", result.User.ID.String(), accessTokenExpiration).Return("", errors.New("token generation failed"))

		handler := NewAuthHandler(mockSvc, mockTM)
		router := setupAuthRouter(handler)

		reqBody := request.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})
}

func TestAuthHandler_OAuthGoogle(t *testing.T) {
	t.Run("新規ユーザーの OAuth 認証が成功する", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		mockTM := new(mockTokenManager)

		user := response.UserResponse{
			ID:          uuid.New(),
			Email:       "oauth@example.com",
			Username:    "oauthuser",
			DisplayName: "OAuth User",
			Role:        "user",
		}
		result := &service.AuthResult{User: user, RefreshToken: "test-refresh-token", IsCreated: true}
		mockSvc.On("OAuthGoogle", mock.Anything, mock.AnythingOfType("request.OAuthGoogleRequest")).Return(result, nil)
		mockTM.On("Generate", user.ID.String(), accessTokenExpiration).Return("test-access-token", nil)

		handler := NewAuthHandler(mockSvc, mockTM)
		router := setupAuthRouter(handler)

		reqBody := request.OAuthGoogleRequest{
			ProviderUserID: "google-user-123",
			Email:          "oauth@example.com",
			DisplayName:    "OAuth User",
			AccessToken:    "access-token",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/auth/oauth/google", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockSvc.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})

	t.Run("既存ユーザーの OAuth 認証が成功する", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		mockTM := new(mockTokenManager)

		user := response.UserResponse{
			ID:          uuid.New(),
			Email:       "existing@example.com",
			Username:    "existinguser",
			DisplayName: "Existing User",
			Role:        "user",
		}
		result := &service.AuthResult{User: user, RefreshToken: "test-refresh-token", IsCreated: false}
		mockSvc.On("OAuthGoogle", mock.Anything, mock.AnythingOfType("request.OAuthGoogleRequest")).Return(result, nil)
		mockTM.On("Generate", user.ID.String(), accessTokenExpiration).Return("test-access-token", nil)

		handler := NewAuthHandler(mockSvc, mockTM)
		router := setupAuthRouter(handler)

		reqBody := request.OAuthGoogleRequest{
			ProviderUserID: "google-user-456",
			Email:          "existing@example.com",
			DisplayName:    "Existing User",
			AccessToken:    "access-token",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/auth/oauth/google", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockSvc.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})

	t.Run("バリデーションエラーの場合は 400 を返す", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		mockTM := new(mockTokenManager)

		handler := NewAuthHandler(mockSvc, mockTM)
		router := setupAuthRouter(handler)

		// 必須フィールドが欠けている
		reqBody := map[string]string{
			"email": "test@example.com",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/auth/oauth/google", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("サービスがエラーを返すとエラーレスポンスを返す", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		mockTM := new(mockTokenManager)

		mockSvc.On("OAuthGoogle", mock.Anything, mock.AnythingOfType("request.OAuthGoogleRequest")).Return(nil, apperror.ErrInternal.WithMessage("OAuth error"))

		handler := NewAuthHandler(mockSvc, mockTM)
		router := setupAuthRouter(handler)

		reqBody := request.OAuthGoogleRequest{
			ProviderUserID: "google-user-123",
			Email:          "oauth@example.com",
			DisplayName:    "OAuth User",
			AccessToken:    "access-token",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/auth/oauth/google", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("トークン生成が失敗した場合は 500 を返す", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		mockTM := new(mockTokenManager)

		user := response.UserResponse{
			ID:          uuid.New(),
			Email:       "oauth@example.com",
			Username:    "oauthuser",
			DisplayName: "OAuth User",
			Role:        "user",
		}
		result := &service.AuthResult{User: user, RefreshToken: "test-refresh-token", IsCreated: true}
		mockSvc.On("OAuthGoogle", mock.Anything, mock.AnythingOfType("request.OAuthGoogleRequest")).Return(result, nil)
		mockTM.On("Generate", user.ID.String(), accessTokenExpiration).Return("", errors.New("token generation failed"))

		handler := NewAuthHandler(mockSvc, mockTM)
		router := setupAuthRouter(handler)

		reqBody := request.OAuthGoogleRequest{
			ProviderUserID: "google-user-123",
			Email:          "oauth@example.com",
			DisplayName:    "OAuth User",
			AccessToken:    "access-token",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/auth/oauth/google", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})
}

func TestAuthHandler_RefreshToken(t *testing.T) {
	t.Run("トークンリフレッシュが成功する", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		mockTM := new(mockTokenManager)

		userID := uuid.New()
		refreshResult := &service.RefreshResult{
			UserID:       userID,
			RefreshToken: "new-refresh-token",
		}
		mockSvc.On("RefreshToken", mock.Anything, mock.AnythingOfType("request.RefreshTokenRequest")).Return(refreshResult, nil)
		mockTM.On("Generate", userID.String(), accessTokenExpiration).Return("new-access-token", nil)

		handler := NewAuthHandler(mockSvc, mockTM)
		router := setupAuthRouter(handler)

		reqBody := request.RefreshTokenRequest{
			RefreshToken: "old-refresh-token",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/auth/refresh", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]response.TokenRefreshResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "new-access-token", resp["data"].AccessToken)
		assert.Equal(t, "new-refresh-token", resp["data"].RefreshToken)
		mockSvc.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})

	t.Run("バリデーションエラーの場合は 400 を返す", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		mockTM := new(mockTokenManager)

		handler := NewAuthHandler(mockSvc, mockTM)
		router := setupAuthRouter(handler)

		// refreshToken が空
		reqBody := map[string]string{}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/auth/refresh", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("無効なリフレッシュトークンの場合は 401 を返す", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		mockTM := new(mockTokenManager)

		mockSvc.On("RefreshToken", mock.Anything, mock.AnythingOfType("request.RefreshTokenRequest")).Return(nil, apperror.ErrInvalidRefreshToken)

		handler := NewAuthHandler(mockSvc, mockTM)
		router := setupAuthRouter(handler)

		reqBody := request.RefreshTokenRequest{
			RefreshToken: "invalid-token",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/auth/refresh", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("アクセストークン生成が失敗した場合は 500 を返す", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		mockTM := new(mockTokenManager)

		userID := uuid.New()
		refreshResult := &service.RefreshResult{
			UserID:       userID,
			RefreshToken: "new-refresh-token",
		}
		mockSvc.On("RefreshToken", mock.Anything, mock.AnythingOfType("request.RefreshTokenRequest")).Return(refreshResult, nil)
		mockTM.On("Generate", userID.String(), accessTokenExpiration).Return("", errors.New("token generation failed"))

		handler := NewAuthHandler(mockSvc, mockTM)
		router := setupAuthRouter(handler)

		reqBody := request.RefreshTokenRequest{
			RefreshToken: "old-refresh-token",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/auth/refresh", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
		mockTM.AssertExpectations(t)
	})
}

func TestAuthHandler_Logout(t *testing.T) {
	userID := uuid.New().String()

	t.Run("ログアウトが成功する", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		mockTM := new(mockTokenManager)

		mockSvc.On("Logout", mock.Anything, userID, mock.AnythingOfType("request.LogoutRequest")).Return(nil)

		handler := NewAuthHandler(mockSvc, mockTM)
		router := setupAuthenticatedAuthRouter(handler, userID)

		reqBody := request.LogoutRequest{
			RefreshToken: "refresh-token-to-revoke",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/auth/logout", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("バリデーションエラーの場合は 400 を返す", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		mockTM := new(mockTokenManager)

		handler := NewAuthHandler(mockSvc, mockTM)
		router := setupAuthenticatedAuthRouter(handler, userID)

		// refreshToken が空
		reqBody := map[string]string{}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/auth/logout", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("無効なリフレッシュトークンの場合は 401 を返す", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		mockTM := new(mockTokenManager)

		mockSvc.On("Logout", mock.Anything, userID, mock.AnythingOfType("request.LogoutRequest")).Return(apperror.ErrInvalidRefreshToken)

		handler := NewAuthHandler(mockSvc, mockTM)
		router := setupAuthenticatedAuthRouter(handler, userID)

		reqBody := request.LogoutRequest{
			RefreshToken: "invalid-token",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/auth/logout", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockSvc.AssertExpectations(t)
	})
}

func TestAuthHandler_GetMe(t *testing.T) {
	userID := uuid.New().String()

	t.Run("現在のユーザー情報を取得できる", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		mockTM := new(mockTokenManager)

		me := &response.MeResponse{
			ID:             uuid.MustParse(userID),
			Email:          "test@example.com",
			Username:       "testuser",
			DisplayName:    "Test User",
			Role:           "user",
			HasPassword:    true,
			OAuthProviders: []string{},
			CreatedAt:      time.Now(),
		}
		mockSvc.On("GetMe", mock.Anything, userID).Return(me, nil)

		handler := NewAuthHandler(mockSvc, mockTM)
		router := setupAuthenticatedAuthRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]response.MeResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "test@example.com", resp["data"].Email)
		mockSvc.AssertExpectations(t)
	})

	t.Run("未認証の場合は 401 を返す", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		mockTM := new(mockTokenManager)

		handler := NewAuthHandler(mockSvc, mockTM)
		router := setupAuthRouter(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("ユーザーが見つからない場合は 404 を返す", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		mockTM := new(mockTokenManager)

		mockSvc.On("GetMe", mock.Anything, userID).Return(nil, apperror.ErrNotFound.WithMessage("User not found"))

		handler := NewAuthHandler(mockSvc, mockTM)
		router := setupAuthenticatedAuthRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("サービスがエラーを返すとエラーレスポンスを返す", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		mockTM := new(mockTokenManager)

		mockSvc.On("GetMe", mock.Anything, userID).Return(nil, apperror.ErrInternal.WithMessage("Database error"))

		handler := NewAuthHandler(mockSvc, mockTM)
		router := setupAuthenticatedAuthRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})
}
