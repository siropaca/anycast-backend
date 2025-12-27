package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "test-secret-key"

// テスト用の有効な JWT トークンを生成する
func generateTestToken(t *testing.T, userID string, expiresAt time.Time) string {
	t.Helper()

	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID: userID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(testSecret))
	require.NoError(t, err)

	return tokenString
}

func setupRouter(jwtSecret string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Auth(jwtSecret))
	r.GET("/test", func(c *gin.Context) {
		userID, _ := GetUserID(c)
		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})
	return r
}

func TestAuth(t *testing.T) {
	t.Run("Authorization ヘッダーがない場合は 401 を返す", func(t *testing.T) {
		router := setupRouter(testSecret)
		req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "UNAUTHORIZED")
	})

	t.Run("Bearer プレフィックスがない場合は 401 を返す", func(t *testing.T) {
		router := setupRouter(testSecret)
		req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
		req.Header.Set("Authorization", "InvalidToken")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "UNAUTHORIZED")
	})

	t.Run("無効なトークンの場合は 401 を返す", func(t *testing.T) {
		router := setupRouter(testSecret)
		req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
		req.Header.Set("Authorization", "Bearer invalid-token")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "UNAUTHORIZED")
	})

	t.Run("期限切れのトークンの場合は 401 を返す", func(t *testing.T) {
		router := setupRouter(testSecret)
		expiredToken := generateTestToken(t, "user-123", time.Now().Add(-1*time.Hour))
		req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
		req.Header.Set("Authorization", "Bearer "+expiredToken)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "UNAUTHORIZED")
	})

	t.Run("異なるシークレットで署名されたトークンの場合は 401 を返す", func(t *testing.T) {
		router := setupRouter(testSecret)

		// 異なるシークレットでトークンを生成
		claims := &Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			},
			UserID: "user-123",
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		wrongToken, err := token.SignedString([]byte("wrong-secret"))
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
		req.Header.Set("Authorization", "Bearer "+wrongToken)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("有効なトークンの場合は次のハンドラーが呼ばれる", func(t *testing.T) {
		router := setupRouter(testSecret)
		validToken := generateTestToken(t, "user-123", time.Now().Add(1*time.Hour))
		req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
		req.Header.Set("Authorization", "Bearer "+validToken)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "user-123")
	})

	t.Run("Bearer の大文字小文字を区別しない", func(t *testing.T) {
		router := setupRouter(testSecret)
		validToken := generateTestToken(t, "user-456", time.Now().Add(1*time.Hour))
		req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
		req.Header.Set("Authorization", "bearer "+validToken)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "user-456")
	})
}

func TestGetUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("コンテキストにユーザー ID がある場合は取得できる", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set(string(UserIDKey), "user-123")

		userID, ok := GetUserID(c)

		assert.True(t, ok)
		assert.Equal(t, "user-123", userID)
	})

	t.Run("コンテキストにユーザー ID がない場合は空文字と false を返す", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		userID, ok := GetUserID(c)

		assert.False(t, ok)
		assert.Empty(t, userID)
	})

	t.Run("ユーザー ID が文字列でない場合は空文字と false を返す", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set(string(UserIDKey), 123) // 文字列ではなく int

		userID, ok := GetUserID(c)

		assert.False(t, ok)
		assert.Empty(t, userID)
	})
}
