package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// ScriptJobService のモック
type mockScriptJobService struct {
	mock.Mock
}

func (m *mockScriptJobService) CreateJob(ctx context.Context, userID, channelID, episodeID string, req request.GenerateScriptAsyncRequest) (*response.ScriptJobResponse, error) {
	args := m.Called(ctx, userID, channelID, episodeID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.ScriptJobResponse), args.Error(1)
}

func (m *mockScriptJobService) GetJob(ctx context.Context, userID, jobID string) (*response.ScriptJobResponse, error) {
	args := m.Called(ctx, userID, jobID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.ScriptJobResponse), args.Error(1)
}

func (m *mockScriptJobService) ListMyJobs(ctx context.Context, userID string, filter repository.ScriptJobFilter) (*response.ScriptJobListResponse, error) {
	args := m.Called(ctx, userID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.ScriptJobListResponse), args.Error(1)
}

func (m *mockScriptJobService) ExecuteJob(ctx context.Context, jobID string) error {
	args := m.Called(ctx, jobID)
	return args.Error(0)
}

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

		handler := NewWorkerHandler(mockSvc, new(mockScriptJobService))
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
		handler := NewWorkerHandler(mockSvc, new(mockScriptJobService))
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

		handler := NewWorkerHandler(mockSvc, new(mockScriptJobService))
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

		handler := NewWorkerHandler(mockSvc, new(mockScriptJobService))
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
		handler := NewWorkerHandler(mockSvc, new(mockScriptJobService))
		assert.NotNil(t, handler)
	})
}
