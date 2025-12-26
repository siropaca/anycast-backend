package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/siropaca/anycast-backend/internal/apperror"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestSuccess(t *testing.T) {
	t.Run("成功レスポンスを返す", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		data := map[string]string{"message": "hello"}
		Success(c, http.StatusOK, data)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "hello", resp["data"]["message"])
	})

	t.Run("201 Created でデータを返す", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		data := map[string]int{"id": 123}
		Success(c, http.StatusCreated, data)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("nil データでも成功レスポンスを返す", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		Success(c, http.StatusOK, nil)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Nil(t, resp["data"])
	})
}

func TestError(t *testing.T) {
	t.Run("AppError をエラーレスポンスに変換する", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		appErr := apperror.New("TEST_ERROR", "test message", http.StatusBadRequest)
		Error(c, appErr)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp map[string]map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "TEST_ERROR", resp["error"]["code"])
		assert.Equal(t, "test message", resp["error"]["message"])
	})

	t.Run("AppError の Details を含める", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		details := map[string]string{"field": "name", "reason": "required"}
		appErr := apperror.New("VALIDATION_ERROR", "validation failed", http.StatusBadRequest).WithDetails(details)
		Error(c, appErr)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp map[string]map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.NotNil(t, resp["error"]["details"])
	})

	t.Run("Details が nil の場合は含めない", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		appErr := apperror.New("TEST_ERROR", "test message", http.StatusBadRequest)
		Error(c, appErr)

		var resp map[string]map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		_, exists := resp["error"]["details"]
		assert.False(t, exists)
	})

	t.Run("未知のエラーは 500 Internal Server Error を返す", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		err := errors.New("unexpected error")
		Error(c, err)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var resp map[string]map[string]any
		jsonErr := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, jsonErr)
		assert.Equal(t, "INTERNAL_ERROR", resp["error"]["code"])
		assert.Equal(t, "Internal server error", resp["error"]["message"])
	})

	t.Run("ラップされた AppError も正しく処理する", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		innerErr := errors.New("db connection failed")
		appErr := apperror.Wrap(innerErr, "DB_ERROR", "database error", http.StatusServiceUnavailable)
		Error(c, appErr)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)

		var resp map[string]map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "DB_ERROR", resp["error"]["code"])
	})
}
