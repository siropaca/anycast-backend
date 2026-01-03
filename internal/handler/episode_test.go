package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/middleware"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// EpisodeService のモック
type mockEpisodeService struct {
	mock.Mock
}

func (m *mockEpisodeService) ListMyChannelEpisodes(ctx context.Context, userID, channelID string, filter repository.EpisodeFilter) (*response.EpisodeListWithPaginationResponse, error) {
	args := m.Called(ctx, userID, channelID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.EpisodeListWithPaginationResponse), args.Error(1)
}

// テスト用のルーターをセットアップする
func setupEpisodeRouter(h *EpisodeHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/me/channels/:channelId/episodes", h.ListMyChannelEpisodes)
	return r
}

// 認証済みルーターをセットアップする
func setupAuthenticatedEpisodeRouter(h *EpisodeHandler, userID string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(string(middleware.UserIDKey), userID)
		c.Next()
	})
	r.GET("/me/channels/:channelId/episodes", h.ListMyChannelEpisodes)
	return r
}

// テスト用のエピソードレスポンスを生成する
func createTestEpisodeResponse() response.EpisodeResponse {
	now := time.Now()
	description := "Test Description"
	return response.EpisodeResponse{
		ID:           uuid.New(),
		Title:        "Test Episode",
		Description:  &description,
		ScriptPrompt: "Test Script Prompt",
		FullAudio: &response.AudioResponse{
			ID:         uuid.New(),
			URL:        "https://example.com/audio.mp3",
			DurationMs: 180000,
		},
		PublishedAt: &now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func TestEpisodeHandler_ListMyChannelEpisodes(t *testing.T) {
	userID := uuid.New().String()
	channelID := uuid.New().String()

	t.Run("エピソード一覧を取得できる", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		episodes := []response.EpisodeResponse{createTestEpisodeResponse()}
		result := &response.EpisodeListWithPaginationResponse{
			Data:       episodes,
			Pagination: response.PaginationResponse{Total: 1, Limit: 20, Offset: 0},
		}
		mockSvc.On("ListMyChannelEpisodes", mock.Anything, userID, channelID, mock.AnythingOfType("repository.EpisodeFilter")).Return(result, nil)

		handler := NewEpisodeHandler(mockSvc)
		router := setupAuthenticatedEpisodeRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me/channels/"+channelID+"/episodes", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp response.EpisodeListWithPaginationResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Len(t, resp.Data, 1)
		mockSvc.AssertExpectations(t)
	})

	t.Run("空のエピソード一覧を取得できる", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		result := &response.EpisodeListWithPaginationResponse{
			Data:       []response.EpisodeResponse{},
			Pagination: response.PaginationResponse{Total: 0, Limit: 20, Offset: 0},
		}
		mockSvc.On("ListMyChannelEpisodes", mock.Anything, userID, channelID, mock.AnythingOfType("repository.EpisodeFilter")).Return(result, nil)

		handler := NewEpisodeHandler(mockSvc)
		router := setupAuthenticatedEpisodeRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me/channels/"+channelID+"/episodes", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp response.EpisodeListWithPaginationResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Empty(t, resp.Data)
		mockSvc.AssertExpectations(t)
	})

	t.Run("クエリパラメータでフィルタできる", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		result := &response.EpisodeListWithPaginationResponse{
			Data:       []response.EpisodeResponse{},
			Pagination: response.PaginationResponse{Total: 0, Limit: 10, Offset: 5},
		}
		mockSvc.On("ListMyChannelEpisodes", mock.Anything, userID, channelID, mock.MatchedBy(func(f repository.EpisodeFilter) bool {
			return f.Limit == 10 && f.Offset == 5 && *f.Status == "published"
		})).Return(result, nil)

		handler := NewEpisodeHandler(mockSvc)
		router := setupAuthenticatedEpisodeRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me/channels/"+channelID+"/episodes?status=published&limit=10&offset=5", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("チャンネルが見つからない場合は 404 を返す", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		mockSvc.On("ListMyChannelEpisodes", mock.Anything, userID, channelID, mock.AnythingOfType("repository.EpisodeFilter")).Return(nil, apperror.ErrNotFound.WithMessage("Channel not found"))

		handler := NewEpisodeHandler(mockSvc)
		router := setupAuthenticatedEpisodeRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me/channels/"+channelID+"/episodes", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("権限がない場合は 403 を返す", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		mockSvc.On("ListMyChannelEpisodes", mock.Anything, userID, channelID, mock.AnythingOfType("repository.EpisodeFilter")).Return(nil, apperror.ErrForbidden.WithMessage("You do not have permission"))

		handler := NewEpisodeHandler(mockSvc)
		router := setupAuthenticatedEpisodeRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me/channels/"+channelID+"/episodes", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("サービスがエラーを返すとエラーレスポンスを返す", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		mockSvc.On("ListMyChannelEpisodes", mock.Anything, userID, channelID, mock.AnythingOfType("repository.EpisodeFilter")).Return(nil, apperror.ErrInternal)

		handler := NewEpisodeHandler(mockSvc)
		router := setupAuthenticatedEpisodeRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me/channels/"+channelID+"/episodes", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("未認証の場合は 401 を返す", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		handler := NewEpisodeHandler(mockSvc)
		router := setupEpisodeRouter(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me/channels/"+channelID+"/episodes", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
