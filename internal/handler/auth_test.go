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
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/middleware"
	"github.com/siropaca/anycast-backend/internal/pkg/jwt"
	"github.com/siropaca/anycast-backend/internal/service"
)

// AuthService のモック
type mockAuthService struct {
	mock.Mock
}

func (m *mockAuthService) Register(ctx context.Context, req request.RegisterRequest) (*response.UserResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.UserResponse), args.Error(1)
}

func (m *mockAuthService) Login(ctx context.Context, req request.LoginRequest) (*response.UserResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.UserResponse), args.Error(1)
}

func (m *mockAuthService) OAuthGoogle(ctx context.Context, req request.OAuthGoogleRequest) (*service.AuthResult, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.AuthResult), args.Error(1)
}

func (m *mockAuthService) GetMe(ctx context.Context, userID string) (*response.MeResponse, error) {
	args := m.Called(ctx, userID)
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
	return r
}

// テスト用のユーザーレスポンスを生成する
func createTestUserResponse() *response.UserResponse {
	return &response.UserResponse{
		ID:          uuid.New(),
		Email:       "test@example.com",
		Username:    "testuser",
		DisplayName: "Test User",
	}
}

func TestAuthHandler_Register(t *testing.T) {
	t.Run("ユーザー登録が成功する", func(t *testing.T) {
		mockSvc := new(mockAuthService)
		mockTM := new(mockTokenManager)

		user := createTestUserResponse()
		mockSvc.On("Register", mock.Anything, mock.AnythingOfType("request.RegisterRequest")).Return(user, nil)
		mockTM.On("Generate", user.ID.String(), tokenExpiration).Return("test-token", nil)

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
		assert.Equal(t, "test-token", resp["data"].Token)
		assert.Equal(t, user.Email, resp["data"].User.Email)
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

		user := createTestUserResponse()
		mockSvc.On("Register", mock.Anything, mock.AnythingOfType("request.RegisterRequest")).Return(user, nil)
		mockTM.On("Generate", user.ID.String(), tokenExpiration).Return("", errors.New("token generation failed"))

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

		user := createTestUserResponse()
		mockSvc.On("Login", mock.Anything, mock.AnythingOfType("request.LoginRequest")).Return(user, nil)
		mockTM.On("Generate", user.ID.String(), tokenExpiration).Return("test-token", nil)

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
		assert.Equal(t, "test-token", resp["data"].Token)
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

		user := createTestUserResponse()
		mockSvc.On("Login", mock.Anything, mock.AnythingOfType("request.LoginRequest")).Return(user, nil)
		mockTM.On("Generate", user.ID.String(), tokenExpiration).Return("", errors.New("token generation failed"))

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
		}
		result := &service.AuthResult{User: user, IsCreated: true}
		mockSvc.On("OAuthGoogle", mock.Anything, mock.AnythingOfType("request.OAuthGoogleRequest")).Return(result, nil)
		mockTM.On("Generate", user.ID.String(), tokenExpiration).Return("test-token", nil)

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
		}
		result := &service.AuthResult{User: user, IsCreated: false}
		mockSvc.On("OAuthGoogle", mock.Anything, mock.AnythingOfType("request.OAuthGoogleRequest")).Return(result, nil)
		mockTM.On("Generate", user.ID.String(), tokenExpiration).Return("test-token", nil)

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
		}
		result := &service.AuthResult{User: user, IsCreated: true}
		mockSvc.On("OAuthGoogle", mock.Anything, mock.AnythingOfType("request.OAuthGoogleRequest")).Return(result, nil)
		mockTM.On("Generate", user.ID.String(), tokenExpiration).Return("", errors.New("token generation failed"))

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
