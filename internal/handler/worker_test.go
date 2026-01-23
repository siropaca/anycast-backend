package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// テスト用のルーターをセットアップする
func setupWorkerRouter(h *WorkerHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/internal/worker/audio", h.ProcessAudioJob)
	return r
}

func TestWorkerHandler_ProcessAudioJob(t *testing.T) {
	jobID := uuid.New().String()

	t.Run("ジョブを正常に処理できる", func(t *testing.T) {
		mockSvc := new(mockAudioJobService)
		mockSvc.On("ExecuteJob", mock.Anything, jobID).Return(nil)

		handler := NewWorkerHandler(mockSvc)
		router := setupWorkerRouter(handler)

		payload := AudioJobPayload{JobID: jobID}
		body, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/internal/worker/audio", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "completed", resp["status"])
		assert.Equal(t, jobID, resp["job_id"])
		mockSvc.AssertExpectations(t)
	})

	t.Run("jobId が指定されていない場合は 400 を返す", func(t *testing.T) {
		mockSvc := new(mockAudioJobService)
		handler := NewWorkerHandler(mockSvc)
		router := setupWorkerRouter(handler)

		payload := map[string]string{}
		body, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/internal/worker/audio", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockSvc.AssertNotCalled(t, "ExecuteJob")
	})

	t.Run("リトライ可能なエラーの場合はエラーレスポンスを返す", func(t *testing.T) {
		mockSvc := new(mockAudioJobService)
		retryableErr := apperror.ErrInternal.WithMessage("temporary error")
		mockSvc.On("ExecuteJob", mock.Anything, jobID).Return(retryableErr)

		handler := NewWorkerHandler(mockSvc)
		router := setupWorkerRouter(handler)

		payload := AudioJobPayload{JobID: jobID}
		body, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/internal/worker/audio", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("非リトライエラーの場合は 200 で失敗ステータスを返す", func(t *testing.T) {
		mockSvc := new(mockAudioJobService)
		// バリデーションエラーはリトライ不可
		nonRetryableErr := apperror.ErrValidation.WithMessage("validation error")
		mockSvc.On("ExecuteJob", mock.Anything, jobID).Return(nonRetryableErr)

		handler := NewWorkerHandler(mockSvc)
		router := setupWorkerRouter(handler)

		payload := AudioJobPayload{JobID: jobID}
		body, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/internal/worker/audio", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "failed", resp["status"])
		assert.Equal(t, jobID, resp["job_id"])
		mockSvc.AssertExpectations(t)
	})
}

func TestNewWorkerHandler(t *testing.T) {
	t.Run("WorkerHandler を作成できる", func(t *testing.T) {
		mockSvc := new(mockAudioJobService)
		handler := NewWorkerHandler(mockSvc)
		assert.NotNil(t, handler)
	})
}
