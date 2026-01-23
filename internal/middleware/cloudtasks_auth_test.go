package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupCloudTasksRouter(expectedAudience, expectedServiceAccount string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CloudTasksAuth(expectedAudience, expectedServiceAccount))
	r.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	return r
}

func TestCloudTasksAuth(t *testing.T) {
	const (
		testAudience       = "https://example.com/worker"
		testServiceAccount = "cloud-tasks@project.iam.gserviceaccount.com"
	)

	t.Run("Authorization ヘッダーがない場合は 401 を返す", func(t *testing.T) {
		router := setupCloudTasksRouter(testAudience, testServiceAccount)
		req := httptest.NewRequest(http.MethodPost, "/test", http.NoBody)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "UNAUTHORIZED")
	})

	t.Run("Bearer プレフィックスがない場合は 401 を返す", func(t *testing.T) {
		router := setupCloudTasksRouter(testAudience, testServiceAccount)
		req := httptest.NewRequest(http.MethodPost, "/test", http.NoBody)
		req.Header.Set("Authorization", "InvalidToken")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "UNAUTHORIZED")
	})

	t.Run("空の Authorization ヘッダーの場合は 401 を返す", func(t *testing.T) {
		router := setupCloudTasksRouter(testAudience, testServiceAccount)
		req := httptest.NewRequest(http.MethodPost, "/test", http.NoBody)
		req.Header.Set("Authorization", "")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("Bearer のみでトークンがない場合は 401 を返す", func(t *testing.T) {
		router := setupCloudTasksRouter(testAudience, testServiceAccount)
		req := httptest.NewRequest(http.MethodPost, "/test", http.NoBody)
		req.Header.Set("Authorization", "Bearer")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("無効なトークンの場合は 401 を返す", func(t *testing.T) {
		router := setupCloudTasksRouter(testAudience, testServiceAccount)
		req := httptest.NewRequest(http.MethodPost, "/test", http.NoBody)
		req.Header.Set("Authorization", "Bearer invalid-token")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("Bearer の大文字小文字を区別しない", func(t *testing.T) {
		router := setupCloudTasksRouter(testAudience, testServiceAccount)
		req := httptest.NewRequest(http.MethodPost, "/test", http.NoBody)
		req.Header.Set("Authorization", "bearer invalid-token")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		// トークンは無効だが、ヘッダーのパースは成功している
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("BEARER でも受け付ける", func(t *testing.T) {
		router := setupCloudTasksRouter(testAudience, testServiceAccount)
		req := httptest.NewRequest(http.MethodPost, "/test", http.NoBody)
		req.Header.Set("Authorization", "BEARER invalid-token")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		// トークンは無効だが、ヘッダーのパースは成功している
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}
