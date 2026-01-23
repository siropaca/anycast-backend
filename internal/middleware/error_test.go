package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
)

// mockHandler はログ出力をキャプチャするためのモックハンドラー
type mockHandler struct {
	mu      sync.Mutex
	records []slog.Record
}

func (h *mockHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func (h *mockHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.records = append(h.records, r)
	return nil
}

func (h *mockHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

func (h *mockHandler) WithGroup(_ string) slog.Handler {
	return h
}

func (h *mockHandler) getRecords() []slog.Record {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.records
}

func setupErrorHandlerRouter(log *slog.Logger) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		ctx := logger.WithContext(c.Request.Context(), log)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	r.Use(ErrorHandler())
	return r
}

func TestErrorHandler(t *testing.T) {
	t.Run("エラーがない場合は何も処理しない", func(t *testing.T) {
		handler := &mockHandler{}
		log := slog.New(handler)
		router := setupErrorHandlerRouter(log)
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Empty(t, handler.getRecords())
	})

	t.Run("AppError の場合は詳細をログに出力する", func(t *testing.T) {
		handler := &mockHandler{}
		log := slog.New(handler)
		router := setupErrorHandlerRouter(log)
		router.GET("/test", func(c *gin.Context) {
			appErr := apperror.ErrNotFound.WithError(errors.New("user not found"))
			_ = c.Error(appErr)
			c.JSON(appErr.HTTPStatus, appErr)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		records := handler.getRecords()
		assert.Len(t, records, 1)
		assert.Equal(t, slog.LevelError, records[0].Level)
		assert.Equal(t, "request error", records[0].Message)

		// 属性を確認
		var attrs []slog.Attr
		records[0].Attrs(func(a slog.Attr) bool {
			attrs = append(attrs, a)
			return true
		})
		assert.Len(t, attrs, 3)
		assert.Equal(t, "code", attrs[0].Key)
		assert.Equal(t, string(apperror.CodeNotFound), attrs[0].Value.String())
		assert.Equal(t, "message", attrs[1].Key)
		assert.Equal(t, "underlying", attrs[2].Key)
	})

	t.Run("通常のエラーの場合は unexpected error としてログ出力する", func(t *testing.T) {
		handler := &mockHandler{}
		log := slog.New(handler)
		router := setupErrorHandlerRouter(log)
		router.GET("/test", func(c *gin.Context) {
			_ = c.Error(errors.New("something went wrong"))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		records := handler.getRecords()
		assert.Len(t, records, 1)
		assert.Equal(t, slog.LevelError, records[0].Level)
		assert.Equal(t, "unexpected error", records[0].Message)

		// 属性を確認
		var attrs []slog.Attr
		records[0].Attrs(func(a slog.Attr) bool {
			attrs = append(attrs, a)
			return true
		})
		assert.Len(t, attrs, 1)
		assert.Equal(t, "error", attrs[0].Key)
	})

	t.Run("複数のエラーがある場合は最後のエラーを処理する", func(t *testing.T) {
		handler := &mockHandler{}
		log := slog.New(handler)
		router := setupErrorHandlerRouter(log)
		router.GET("/test", func(c *gin.Context) {
			_ = c.Error(errors.New("first error"))
			_ = c.Error(apperror.ErrValidation)
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation error"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		records := handler.getRecords()
		assert.Len(t, records, 1)
		assert.Equal(t, "request error", records[0].Message)

		// 最後のエラー（AppError）が処理されていることを確認
		var attrs []slog.Attr
		records[0].Attrs(func(a slog.Attr) bool {
			attrs = append(attrs, a)
			return true
		})
		assert.Equal(t, "code", attrs[0].Key)
		assert.Equal(t, string(apperror.CodeValidation), attrs[0].Value.String())
	})
}
