package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/middleware"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// AudioJobService のモック
type mockAudioJobService struct {
	mock.Mock
}

func (m *mockAudioJobService) CreateJob(ctx context.Context, userID, channelID, episodeID string, req request.GenerateAudioAsyncRequest) (*response.AudioJobResponse, error) {
	args := m.Called(ctx, userID, channelID, episodeID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.AudioJobResponse), args.Error(1)
}

func (m *mockAudioJobService) GetJob(ctx context.Context, userID, jobID string) (*response.AudioJobResponse, error) {
	args := m.Called(ctx, userID, jobID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.AudioJobResponse), args.Error(1)
}

func (m *mockAudioJobService) ListMyJobs(ctx context.Context, userID string, filter repository.AudioJobFilter) (*response.AudioJobListResponse, error) {
	args := m.Called(ctx, userID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.AudioJobListResponse), args.Error(1)
}

func (m *mockAudioJobService) ExecuteJob(ctx context.Context, jobID string) error {
	args := m.Called(ctx, jobID)
	return args.Error(0)
}

func setupAudioJobRouter(service *mockAudioJobService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewAudioJobHandler(service)

	// 認証済みユーザーをシミュレートするミドルウェア
	authMiddleware := func(userID string) gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set(string(middleware.UserIDKey), userID)
			c.Next()
		}
	}

	r.POST("/me/channels/:channelId/episodes/:episodeId/audio/generate-async", authMiddleware("user-123"), handler.GenerateAudioAsync)
	r.GET("/audio-jobs/:jobId", authMiddleware("user-123"), handler.GetAudioJob)
	r.GET("/me/audio-jobs", authMiddleware("user-123"), handler.ListMyAudioJobs)

	return r
}

func TestAudioJobHandler_GenerateAudioAsync(t *testing.T) {
	channelID := uuid.New()
	episodeID := uuid.New()
	jobID := uuid.New()

	t.Run("音声生成ジョブを作成できる", func(t *testing.T) {
		mockService := new(mockAudioJobService)
		jobResponse := &response.AudioJobResponse{
			ID:        jobID,
			EpisodeID: episodeID,
			Status:    "pending",
			Progress:  0,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		mockService.On("CreateJob", mock.Anything, "user-123", channelID.String(), episodeID.String(), mock.Anything).Return(jobResponse, nil)

		router := setupAudioJobRouter(mockService)
		req := httptest.NewRequest(http.MethodPost, "/me/channels/"+channelID.String()+"/episodes/"+episodeID.String()+"/audio/generate-async", strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusAccepted, rec.Code)

		var resp map[string]response.AudioJobResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "pending", resp["data"].Status)
		mockService.AssertExpectations(t)
	})

	t.Run("オプションパラメータ付きでジョブを作成できる", func(t *testing.T) {
		mockService := new(mockAudioJobService)
		jobResponse := &response.AudioJobResponse{
			ID:        jobID,
			EpisodeID: episodeID,
			Status:    "pending",
			Progress:  0,
		}
		mockService.On("CreateJob", mock.Anything, "user-123", channelID.String(), episodeID.String(), mock.Anything).Return(jobResponse, nil)

		router := setupAudioJobRouter(mockService)
		body := `{"voiceStyle":"warm tone","bgmVolumeDb":-20,"fadeOutMs":5000}`
		req := httptest.NewRequest(http.MethodPost, "/me/channels/"+channelID.String()+"/episodes/"+episodeID.String()+"/audio/generate-async", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusAccepted, rec.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("channelId が空の場合はエラーを返す", func(t *testing.T) {
		mockService := new(mockAudioJobService)

		gin.SetMode(gin.TestMode)
		r := gin.New()
		handler := NewAudioJobHandler(mockService)
		r.POST("/me/channels/:channelId/episodes/:episodeId/audio/generate-async", func(c *gin.Context) {
			c.Set(string(middleware.UserIDKey), "user-123")
			c.Params = gin.Params{{Key: "channelId", Value: ""}, {Key: "episodeId", Value: "ep-123"}}
			handler.GenerateAudioAsync(c)
		})

		req := httptest.NewRequest(http.MethodPost, "/me/channels//episodes/ep-123/audio/generate-async", http.NoBody)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("サービスがエラーを返すとエラーを返す", func(t *testing.T) {
		mockService := new(mockAudioJobService)
		mockService.On("CreateJob", mock.Anything, "user-123", channelID.String(), episodeID.String(), mock.Anything).Return(nil, apperror.ErrValidation.WithMessage("Validation error"))

		router := setupAudioJobRouter(mockService)
		req := httptest.NewRequest(http.MethodPost, "/me/channels/"+channelID.String()+"/episodes/"+episodeID.String()+"/audio/generate-async", strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		mockService.AssertExpectations(t)
	})
}

func TestAudioJobHandler_GetAudioJob(t *testing.T) {
	jobID := uuid.New()
	episodeID := uuid.New()

	t.Run("ジョブを取得できる", func(t *testing.T) {
		mockService := new(mockAudioJobService)
		jobResponse := &response.AudioJobResponse{
			ID:        jobID,
			EpisodeID: episodeID,
			Status:    "completed",
			Progress:  100,
		}
		mockService.On("GetJob", mock.Anything, "user-123", jobID.String()).Return(jobResponse, nil)

		router := setupAudioJobRouter(mockService)
		req := httptest.NewRequest(http.MethodGet, "/audio-jobs/"+jobID.String(), http.NoBody)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]response.AudioJobResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "completed", resp["data"].Status)
		mockService.AssertExpectations(t)
	})

	t.Run("存在しないジョブは 404 を返す", func(t *testing.T) {
		mockService := new(mockAudioJobService)
		mockService.On("GetJob", mock.Anything, "user-123", jobID.String()).Return(nil, apperror.ErrNotFound.WithMessage("Job not found"))

		router := setupAudioJobRouter(mockService)
		req := httptest.NewRequest(http.MethodGet, "/audio-jobs/"+jobID.String(), http.NoBody)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("他ユーザーのジョブは 403 を返す", func(t *testing.T) {
		mockService := new(mockAudioJobService)
		mockService.On("GetJob", mock.Anything, "user-123", jobID.String()).Return(nil, apperror.ErrForbidden.WithMessage("Access denied"))

		router := setupAudioJobRouter(mockService)
		req := httptest.NewRequest(http.MethodGet, "/audio-jobs/"+jobID.String(), http.NoBody)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		mockService.AssertExpectations(t)
	})
}

func TestAudioJobHandler_ListMyAudioJobs(t *testing.T) {
	jobID1 := uuid.New()
	jobID2 := uuid.New()
	episodeID := uuid.New()

	t.Run("ジョブ一覧を取得できる", func(t *testing.T) {
		mockService := new(mockAudioJobService)
		listResponse := &response.AudioJobListResponse{
			Data: []response.AudioJobResponse{
				{ID: jobID1, EpisodeID: episodeID, Status: "completed", Progress: 100},
				{ID: jobID2, EpisodeID: episodeID, Status: "pending", Progress: 0},
			},
		}
		mockService.On("ListMyJobs", mock.Anything, "user-123", mock.Anything).Return(listResponse, nil)

		router := setupAudioJobRouter(mockService)
		req := httptest.NewRequest(http.MethodGet, "/me/audio-jobs", http.NoBody)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp response.AudioJobListResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Len(t, resp.Data, 2)
		mockService.AssertExpectations(t)
	})

	t.Run("ステータスでフィルタできる", func(t *testing.T) {
		mockService := new(mockAudioJobService)
		listResponse := &response.AudioJobListResponse{
			Data: []response.AudioJobResponse{
				{ID: jobID1, EpisodeID: episodeID, Status: "pending", Progress: 0},
			},
		}
		mockService.On("ListMyJobs", mock.Anything, "user-123", mock.Anything).Return(listResponse, nil)

		router := setupAudioJobRouter(mockService)
		req := httptest.NewRequest(http.MethodGet, "/me/audio-jobs?status=pending", http.NoBody)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("空の一覧を取得できる", func(t *testing.T) {
		mockService := new(mockAudioJobService)
		listResponse := &response.AudioJobListResponse{
			Data: []response.AudioJobResponse{},
		}
		mockService.On("ListMyJobs", mock.Anything, "user-123", mock.Anything).Return(listResponse, nil)

		router := setupAudioJobRouter(mockService)
		req := httptest.NewRequest(http.MethodGet, "/me/audio-jobs", http.NoBody)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp response.AudioJobListResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Empty(t, resp.Data)
		mockService.AssertExpectations(t)
	})

	t.Run("無効なステータスはバリデーションエラー", func(t *testing.T) {
		mockService := new(mockAudioJobService)

		router := setupAudioJobRouter(mockService)
		req := httptest.NewRequest(http.MethodGet, "/me/audio-jobs?status=invalid", http.NoBody)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}
