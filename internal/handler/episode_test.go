package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
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
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
	"github.com/siropaca/anycast-backend/internal/service"
)

// EpisodeService のモック
type mockEpisodeService struct {
	mock.Mock
}

func (m *mockEpisodeService) ListChannelEpisodes(ctx context.Context, userID, channelID string, filter repository.EpisodeFilter) (*response.EpisodeListWithPaginationResponse, error) {
	args := m.Called(ctx, userID, channelID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.EpisodeListWithPaginationResponse), args.Error(1)
}

func (m *mockEpisodeService) ListMyChannelEpisodes(ctx context.Context, userID, channelID string, filter repository.EpisodeFilter) (*response.EpisodeListWithPaginationResponse, error) {
	args := m.Called(ctx, userID, channelID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.EpisodeListWithPaginationResponse), args.Error(1)
}

func (m *mockEpisodeService) GetEpisode(ctx context.Context, userID, channelID, episodeID string) (*response.EpisodeDataResponse, error) {
	args := m.Called(ctx, userID, channelID, episodeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.EpisodeDataResponse), args.Error(1)
}

func (m *mockEpisodeService) GetMyChannelEpisode(ctx context.Context, userID, channelID, episodeID string) (*response.EpisodeDataResponse, error) {
	args := m.Called(ctx, userID, channelID, episodeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.EpisodeDataResponse), args.Error(1)
}

func (m *mockEpisodeService) CreateEpisode(ctx context.Context, userID, channelID, title, description string, artworkImageID *string) (*response.EpisodeResponse, error) {
	args := m.Called(ctx, userID, channelID, title, description, artworkImageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.EpisodeResponse), args.Error(1)
}

func (m *mockEpisodeService) UpdateEpisode(ctx context.Context, userID, channelID, episodeID string, req request.UpdateEpisodeRequest) (*response.EpisodeDataResponse, error) {
	args := m.Called(ctx, userID, channelID, episodeID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.EpisodeDataResponse), args.Error(1)
}

func (m *mockEpisodeService) DeleteEpisode(ctx context.Context, userID, channelID, episodeID string) error {
	args := m.Called(ctx, userID, channelID, episodeID)
	return args.Error(0)
}

func (m *mockEpisodeService) PublishEpisode(ctx context.Context, userID, channelID, episodeID string, publishedAt *string) (*response.EpisodeDataResponse, error) {
	args := m.Called(ctx, userID, channelID, episodeID, publishedAt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.EpisodeDataResponse), args.Error(1)
}

func (m *mockEpisodeService) UnpublishEpisode(ctx context.Context, userID, channelID, episodeID string) (*response.EpisodeDataResponse, error) {
	args := m.Called(ctx, userID, channelID, episodeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.EpisodeDataResponse), args.Error(1)
}

func (m *mockEpisodeService) IncrementPlayCount(ctx context.Context, episodeID string) error {
	args := m.Called(ctx, episodeID)
	return args.Error(0)
}

func (m *mockEpisodeService) SetBgm(ctx context.Context, userID, channelID, episodeID string, req request.SetEpisodeBgmRequest) (*response.EpisodeDataResponse, error) {
	args := m.Called(ctx, userID, channelID, episodeID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.EpisodeDataResponse), args.Error(1)
}

func (m *mockEpisodeService) DeleteBgm(ctx context.Context, userID, channelID, episodeID string) (*response.EpisodeDataResponse, error) {
	args := m.Called(ctx, userID, channelID, episodeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.EpisodeDataResponse), args.Error(1)
}

func (m *mockEpisodeService) DeleteAudio(ctx context.Context, userID, channelID, episodeID string) error {
	args := m.Called(ctx, userID, channelID, episodeID)
	return args.Error(0)
}

func (m *mockEpisodeService) UploadAudio(ctx context.Context, userID, channelID, episodeID string, input service.UploadAudioInput) (*response.EpisodeDataResponse, error) {
	args := m.Called(ctx, userID, channelID, episodeID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.EpisodeDataResponse), args.Error(1)
}

// テスト用のルーターをセットアップする
func setupEpisodeRouter(h *EpisodeHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/me/channels/:channelId/episodes", h.ListMyChannelEpisodes)
	r.GET("/me/channels/:channelId/episodes/:episodeId", h.GetMyChannelEpisode)
	r.GET("/channels/:channelId/episodes", h.ListChannelEpisodes)
	r.PATCH("/channels/:channelId/episodes/:episodeId", h.UpdateEpisode)
	r.DELETE("/channels/:channelId/episodes/:episodeId", h.DeleteEpisode)
	r.POST("/channels/:channelId/episodes/:episodeId/publish", h.PublishEpisode)
	r.POST("/channels/:channelId/episodes/:episodeId/unpublish", h.UnpublishEpisode)
	r.PUT("/channels/:channelId/episodes/:episodeId/audio", h.UploadAudio)
	r.DELETE("/channels/:channelId/episodes/:episodeId/audio", h.DeleteAudio)
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
	r.GET("/me/channels/:channelId/episodes/:episodeId", h.GetMyChannelEpisode)
	r.GET("/channels/:channelId/episodes", h.ListChannelEpisodes)
	r.POST("/channels/:channelId/episodes", h.CreateEpisode)
	r.PATCH("/channels/:channelId/episodes/:episodeId", h.UpdateEpisode)
	r.DELETE("/channels/:channelId/episodes/:episodeId", h.DeleteEpisode)
	r.POST("/channels/:channelId/episodes/:episodeId/publish", h.PublishEpisode)
	r.POST("/channels/:channelId/episodes/:episodeId/unpublish", h.UnpublishEpisode)
	r.PUT("/channels/:channelId/episodes/:episodeId/audio", h.UploadAudio)
	r.DELETE("/channels/:channelId/episodes/:episodeId/audio", h.DeleteAudio)
	return r
}

// テスト用のエピソードレスポンスを生成する
func createTestEpisodeResponse() response.EpisodeResponse {
	now := time.Now()
	return response.EpisodeResponse{
		ID:          uuid.New(),
		Title:       "Test Episode",
		Description: "Test Description",
		FullAudio: &response.AudioResponse{
			ID:         uuid.New(),
			URL:        "https://example.com/audio.mp3",
			MimeType:   "audio/mpeg",
			FileSize:   1024000,
			DurationMs: 180000,
		},
		PublishedAt: &now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func TestEpisodeHandler_ListChannelEpisodes(t *testing.T) {
	userID := uuid.New().String()
	channelID := uuid.New().String()

	t.Run("エピソード一覧を取得できる", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		episodes := []response.EpisodeResponse{createTestEpisodeResponse()}
		result := &response.EpisodeListWithPaginationResponse{
			Data:       episodes,
			Pagination: response.PaginationResponse{Total: 1, Limit: 20, Offset: 0},
		}
		mockSvc.On("ListChannelEpisodes", mock.Anything, userID, channelID, mock.AnythingOfType("repository.EpisodeFilter")).Return(result, nil)

		handler := NewEpisodeHandler(mockSvc)
		router := setupAuthenticatedEpisodeRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/channels/"+channelID+"/episodes", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp response.EpisodeListWithPaginationResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Len(t, resp.Data, 1)
		mockSvc.AssertExpectations(t)
	})

	t.Run("ページネーションパラメータを指定できる", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		result := &response.EpisodeListWithPaginationResponse{
			Data:       []response.EpisodeResponse{},
			Pagination: response.PaginationResponse{Total: 0, Limit: 10, Offset: 5},
		}
		mockSvc.On("ListChannelEpisodes", mock.Anything, userID, channelID, mock.MatchedBy(func(f repository.EpisodeFilter) bool {
			return f.Limit == 10 && f.Offset == 5
		})).Return(result, nil)

		handler := NewEpisodeHandler(mockSvc)
		router := setupAuthenticatedEpisodeRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/channels/"+channelID+"/episodes?limit=10&offset=5", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("チャンネルが見つからない場合は 404 を返す", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		mockSvc.On("ListChannelEpisodes", mock.Anything, userID, channelID, mock.AnythingOfType("repository.EpisodeFilter")).Return(nil, apperror.ErrNotFound.WithMessage("チャンネルが見つかりません"))

		handler := NewEpisodeHandler(mockSvc)
		router := setupAuthenticatedEpisodeRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/channels/"+channelID+"/episodes", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("サービスがエラーを返すとエラーレスポンスを返す", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		mockSvc.On("ListChannelEpisodes", mock.Anything, userID, channelID, mock.AnythingOfType("repository.EpisodeFilter")).Return(nil, apperror.ErrInternal)

		handler := NewEpisodeHandler(mockSvc)
		router := setupAuthenticatedEpisodeRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/channels/"+channelID+"/episodes", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("未認証でも公開エピソード一覧を取得できる", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		episodes := []response.EpisodeResponse{createTestEpisodeResponse()}
		result := &response.EpisodeListWithPaginationResponse{
			Data:       episodes,
			Pagination: response.PaginationResponse{Total: 1, Limit: 20, Offset: 0},
		}
		mockSvc.On("ListChannelEpisodes", mock.Anything, "", channelID, mock.AnythingOfType("repository.EpisodeFilter")).Return(result, nil)

		handler := NewEpisodeHandler(mockSvc)
		router := setupEpisodeRouter(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/channels/"+channelID+"/episodes", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockSvc.AssertExpectations(t)
	})
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

func TestEpisodeHandler_GetMyChannelEpisode(t *testing.T) {
	userID := uuid.New().String()
	channelID := uuid.New().String()
	episodeID := uuid.New().String()

	t.Run("自分のチャンネルのエピソードを取得できる", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		episodeResp := createTestEpisodeResponse()
		result := &response.EpisodeDataResponse{Data: episodeResp}
		mockSvc.On("GetMyChannelEpisode", mock.Anything, userID, channelID, episodeID).Return(result, nil)

		handler := NewEpisodeHandler(mockSvc)
		router := setupAuthenticatedEpisodeRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me/channels/"+channelID+"/episodes/"+episodeID, http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp response.EpisodeDataResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.Data.ID)
		mockSvc.AssertExpectations(t)
	})

	t.Run("チャンネルが見つからない場合は 404 を返す", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		mockSvc.On("GetMyChannelEpisode", mock.Anything, userID, channelID, episodeID).Return(nil, apperror.ErrNotFound.WithMessage("Channel not found"))

		handler := NewEpisodeHandler(mockSvc)
		router := setupAuthenticatedEpisodeRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me/channels/"+channelID+"/episodes/"+episodeID, http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("エピソードが見つからない場合は 404 を返す", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		mockSvc.On("GetMyChannelEpisode", mock.Anything, userID, channelID, episodeID).Return(nil, apperror.ErrNotFound.WithMessage("Episode not found"))

		handler := NewEpisodeHandler(mockSvc)
		router := setupAuthenticatedEpisodeRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me/channels/"+channelID+"/episodes/"+episodeID, http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("権限がない場合は 403 を返す", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		mockSvc.On("GetMyChannelEpisode", mock.Anything, userID, channelID, episodeID).Return(nil, apperror.ErrForbidden.WithMessage("You do not have permission"))

		handler := NewEpisodeHandler(mockSvc)
		router := setupAuthenticatedEpisodeRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me/channels/"+channelID+"/episodes/"+episodeID, http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("サービスがエラーを返すとエラーレスポンスを返す", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		mockSvc.On("GetMyChannelEpisode", mock.Anything, userID, channelID, episodeID).Return(nil, apperror.ErrInternal)

		handler := NewEpisodeHandler(mockSvc)
		router := setupAuthenticatedEpisodeRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me/channels/"+channelID+"/episodes/"+episodeID, http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("未認証の場合は 401 を返す", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		handler := NewEpisodeHandler(mockSvc)
		router := setupEpisodeRouter(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me/channels/"+channelID+"/episodes/"+episodeID, http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestEpisodeHandler_CreateEpisode(t *testing.T) {
	userID := uuid.New().String()
	channelID := uuid.New().String()

	t.Run("エピソードを作成できる", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		result := &response.EpisodeResponse{
			ID:          uuid.New(),
			Title:       "Test Episode",
			Description: "Test Description",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		mockSvc.On("CreateEpisode", mock.Anything, userID, channelID, "Test Episode", "Test Description", (*string)(nil)).Return(result, nil)

		handler := NewEpisodeHandler(mockSvc)
		router := setupAuthenticatedEpisodeRouter(handler, userID)

		body := `{"title":"Test Episode","description":"Test Description"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/channels/"+channelID+"/episodes", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp response.EpisodeDataResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "Test Episode", resp.Data.Title)
		mockSvc.AssertExpectations(t)
	})

	t.Run("必須フィールドが欠けているとバリデーションエラーを返す", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		handler := NewEpisodeHandler(mockSvc)
		router := setupAuthenticatedEpisodeRouter(handler, userID)

		// title が欠けている（title は required）
		body := `{"description":"Test Description"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/channels/"+channelID+"/episodes", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("title と description が空の場合は 400 を返す", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		handler := NewEpisodeHandler(mockSvc)
		router := setupAuthenticatedEpisodeRouter(handler, userID)

		body := `{}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/channels/"+channelID+"/episodes", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("チャンネルが見つからない場合は 404 を返す", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		mockSvc.On("CreateEpisode", mock.Anything, userID, channelID, mock.Anything, mock.Anything, mock.Anything).Return(nil, apperror.ErrNotFound.WithMessage("Channel not found"))

		handler := NewEpisodeHandler(mockSvc)
		router := setupAuthenticatedEpisodeRouter(handler, userID)

		body := `{"title":"Test Episode","description":"Test Description"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/channels/"+channelID+"/episodes", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("権限がない場合は 403 を返す", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		mockSvc.On("CreateEpisode", mock.Anything, userID, channelID, mock.Anything, mock.Anything, mock.Anything).Return(nil, apperror.ErrForbidden.WithMessage("You do not have permission"))

		handler := NewEpisodeHandler(mockSvc)
		router := setupAuthenticatedEpisodeRouter(handler, userID)

		body := `{"title":"Test Episode","description":"Test Description"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/channels/"+channelID+"/episodes", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		mockSvc.AssertExpectations(t)
	})
}

func TestEpisodeHandler_UpdateEpisode(t *testing.T) {
	userID := uuid.New().String()
	channelID := uuid.New().String()
	episodeID := uuid.New().String()

	t.Run("エピソードを更新できる", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		title := "Updated Title"
		result := &response.EpisodeDataResponse{
			Data: response.EpisodeResponse{
				ID:          uuid.MustParse(episodeID),
				Title:       title,
				Description: "Updated Description",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		}
		mockSvc.On("UpdateEpisode", mock.Anything, userID, channelID, episodeID, mock.AnythingOfType("request.UpdateEpisodeRequest")).Return(result, nil)

		handler := NewEpisodeHandler(mockSvc)
		router := setupAuthenticatedEpisodeRouter(handler, userID)

		body := `{"title":"Updated Title","description":"Updated Description"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("PATCH", "/channels/"+channelID+"/episodes/"+episodeID, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp response.EpisodeDataResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Title", resp.Data.Title)
		mockSvc.AssertExpectations(t)
	})

	t.Run("必須フィールドが欠けているとバリデーションエラーを返す", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		handler := NewEpisodeHandler(mockSvc)
		router := setupAuthenticatedEpisodeRouter(handler, userID)

		// title のみで description が欠けている
		body := `{"title":"Updated Title"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("PATCH", "/channels/"+channelID+"/episodes/"+episodeID, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("エピソードが見つからない場合は 404 を返す", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		mockSvc.On("UpdateEpisode", mock.Anything, userID, channelID, episodeID, mock.AnythingOfType("request.UpdateEpisodeRequest")).Return(nil, apperror.ErrNotFound.WithMessage("Episode not found"))

		handler := NewEpisodeHandler(mockSvc)
		router := setupAuthenticatedEpisodeRouter(handler, userID)

		body := `{"title":"Updated Title","description":"Updated Description"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("PATCH", "/channels/"+channelID+"/episodes/"+episodeID, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("権限がない場合は 403 を返す", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		mockSvc.On("UpdateEpisode", mock.Anything, userID, channelID, episodeID, mock.AnythingOfType("request.UpdateEpisodeRequest")).Return(nil, apperror.ErrForbidden.WithMessage("You do not have permission"))

		handler := NewEpisodeHandler(mockSvc)
		router := setupAuthenticatedEpisodeRouter(handler, userID)

		body := `{"title":"Updated Title","description":"Updated Description"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("PATCH", "/channels/"+channelID+"/episodes/"+episodeID, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("不正な JSON の場合は 400 を返す", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		handler := NewEpisodeHandler(mockSvc)
		router := setupAuthenticatedEpisodeRouter(handler, userID)

		body := `{"title":}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("PATCH", "/channels/"+channelID+"/episodes/"+episodeID, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("未認証の場合は 401 を返す", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		handler := NewEpisodeHandler(mockSvc)
		router := setupEpisodeRouter(handler)

		body := `{"title":"Updated Title","description":"Updated Description"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("PATCH", "/channels/"+channelID+"/episodes/"+episodeID, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestEpisodeHandler_DeleteEpisode(t *testing.T) {
	userID := uuid.New().String()
	channelID := uuid.New().String()
	episodeID := uuid.New().String()

	t.Run("エピソードを削除できる", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		mockSvc.On("DeleteEpisode", mock.Anything, userID, channelID, episodeID).Return(nil)

		handler := NewEpisodeHandler(mockSvc)
		router := setupAuthenticatedEpisodeRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/channels/"+channelID+"/episodes/"+episodeID, http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Empty(t, w.Body.String())
		mockSvc.AssertExpectations(t)
	})

	t.Run("エピソードが見つからない場合は 404 を返す", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		mockSvc.On("DeleteEpisode", mock.Anything, userID, channelID, episodeID).Return(apperror.ErrNotFound.WithMessage("Episode not found"))

		handler := NewEpisodeHandler(mockSvc)
		router := setupAuthenticatedEpisodeRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/channels/"+channelID+"/episodes/"+episodeID, http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("権限がない場合は 403 を返す", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		mockSvc.On("DeleteEpisode", mock.Anything, userID, channelID, episodeID).Return(apperror.ErrForbidden.WithMessage("You do not have permission"))

		handler := NewEpisodeHandler(mockSvc)
		router := setupAuthenticatedEpisodeRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/channels/"+channelID+"/episodes/"+episodeID, http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("未認証の場合は 401 を返す", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		handler := NewEpisodeHandler(mockSvc)
		router := setupEpisodeRouter(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/channels/"+channelID+"/episodes/"+episodeID, http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestEpisodeHandler_PublishEpisode(t *testing.T) {
	userID := uuid.New().String()
	channelID := uuid.New().String()
	episodeID := uuid.New().String()

	t.Run("publishedAt を省略するとエピソードを即時公開できる", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		now := time.Now()
		result := &response.EpisodeDataResponse{
			Data: response.EpisodeResponse{
				ID:          uuid.MustParse(episodeID),
				Title:       "Test Episode",
				PublishedAt: &now,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		}
		mockSvc.On("PublishEpisode", mock.Anything, userID, channelID, episodeID, (*string)(nil)).Return(result, nil)

		handler := NewEpisodeHandler(mockSvc)
		router := setupAuthenticatedEpisodeRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/channels/"+channelID+"/episodes/"+episodeID+"/publish", bytes.NewBufferString("{}"))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp response.EpisodeDataResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.NotNil(t, resp.Data.PublishedAt)
		mockSvc.AssertExpectations(t)
	})

	t.Run("publishedAt を指定するとその日時で公開できる", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		publishedAt := "2025-12-31T23:59:59Z"
		parsedTime, _ := time.Parse(time.RFC3339, publishedAt)
		result := &response.EpisodeDataResponse{
			Data: response.EpisodeResponse{
				ID:          uuid.MustParse(episodeID),
				Title:       "Test Episode",
				PublishedAt: &parsedTime,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		}
		mockSvc.On("PublishEpisode", mock.Anything, userID, channelID, episodeID, &publishedAt).Return(result, nil)

		handler := NewEpisodeHandler(mockSvc)
		router := setupAuthenticatedEpisodeRouter(handler, userID)

		body := `{"publishedAt":"2025-12-31T23:59:59Z"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/channels/"+channelID+"/episodes/"+episodeID+"/publish", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("エピソードが見つからない場合は 404 を返す", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		mockSvc.On("PublishEpisode", mock.Anything, userID, channelID, episodeID, (*string)(nil)).Return(nil, apperror.ErrNotFound.WithMessage("Episode not found"))

		handler := NewEpisodeHandler(mockSvc)
		router := setupAuthenticatedEpisodeRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/channels/"+channelID+"/episodes/"+episodeID+"/publish", bytes.NewBufferString("{}"))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("権限がない場合は 403 を返す", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		mockSvc.On("PublishEpisode", mock.Anything, userID, channelID, episodeID, (*string)(nil)).Return(nil, apperror.ErrForbidden.WithMessage("You do not have permission"))

		handler := NewEpisodeHandler(mockSvc)
		router := setupAuthenticatedEpisodeRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/channels/"+channelID+"/episodes/"+episodeID+"/publish", bytes.NewBufferString("{}"))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("未認証の場合は 401 を返す", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		handler := NewEpisodeHandler(mockSvc)
		router := setupEpisodeRouter(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/channels/"+channelID+"/episodes/"+episodeID+"/publish", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestEpisodeHandler_UnpublishEpisode(t *testing.T) {
	userID := uuid.New().String()
	channelID := uuid.New().String()
	episodeID := uuid.New().String()

	t.Run("エピソードを非公開にできる", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		result := &response.EpisodeDataResponse{
			Data: response.EpisodeResponse{
				ID:          uuid.MustParse(episodeID),
				Title:       "Test Episode",
				PublishedAt: nil,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		}
		mockSvc.On("UnpublishEpisode", mock.Anything, userID, channelID, episodeID).Return(result, nil)

		handler := NewEpisodeHandler(mockSvc)
		router := setupAuthenticatedEpisodeRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/channels/"+channelID+"/episodes/"+episodeID+"/unpublish", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp response.EpisodeDataResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Nil(t, resp.Data.PublishedAt)
		mockSvc.AssertExpectations(t)
	})

	t.Run("エピソードが見つからない場合は 404 を返す", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		mockSvc.On("UnpublishEpisode", mock.Anything, userID, channelID, episodeID).Return(nil, apperror.ErrNotFound.WithMessage("Episode not found"))

		handler := NewEpisodeHandler(mockSvc)
		router := setupAuthenticatedEpisodeRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/channels/"+channelID+"/episodes/"+episodeID+"/unpublish", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("権限がない場合は 403 を返す", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		mockSvc.On("UnpublishEpisode", mock.Anything, userID, channelID, episodeID).Return(nil, apperror.ErrForbidden.WithMessage("You do not have permission"))

		handler := NewEpisodeHandler(mockSvc)
		router := setupAuthenticatedEpisodeRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/channels/"+channelID+"/episodes/"+episodeID+"/unpublish", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("未認証の場合は 401 を返す", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		handler := NewEpisodeHandler(mockSvc)
		router := setupEpisodeRouter(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/channels/"+channelID+"/episodes/"+episodeID+"/unpublish", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestEpisodeHandler_DeleteAudio(t *testing.T) {
	userID := uuid.New().String()
	channelID := uuid.New().String()
	episodeID := uuid.New().String()

	t.Run("エピソードの音声を削除できる", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		mockSvc.On("DeleteAudio", mock.Anything, userID, channelID, episodeID).Return(nil)

		handler := NewEpisodeHandler(mockSvc)
		router := setupAuthenticatedEpisodeRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/channels/"+channelID+"/episodes/"+episodeID+"/audio", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Empty(t, w.Body.String())
		mockSvc.AssertExpectations(t)
	})

	t.Run("未認証の場合は 401 を返す", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		handler := NewEpisodeHandler(mockSvc)
		router := setupEpisodeRouter(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/channels/"+channelID+"/episodes/"+episodeID+"/audio", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("権限がない場合は 403 を返す", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		mockSvc.On("DeleteAudio", mock.Anything, userID, channelID, episodeID).Return(apperror.ErrForbidden.WithMessage("このエピソードの音声削除権限がありません"))

		handler := NewEpisodeHandler(mockSvc)
		router := setupAuthenticatedEpisodeRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/channels/"+channelID+"/episodes/"+episodeID+"/audio", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("エピソードが見つからない場合は 404 を返す", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		mockSvc.On("DeleteAudio", mock.Anything, userID, channelID, episodeID).Return(apperror.ErrNotFound.WithMessage("エピソードが見つかりません"))

		handler := NewEpisodeHandler(mockSvc)
		router := setupAuthenticatedEpisodeRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/channels/"+channelID+"/episodes/"+episodeID+"/audio", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockSvc.AssertExpectations(t)
	})
}

// createEpisodeAudioMultipartForm はテスト用のマルチパートフォームを作成する
func createEpisodeAudioMultipartForm(t *testing.T, filename string, content []byte) (body *bytes.Buffer, formContentType string) {
	t.Helper()
	body = new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	assert.NoError(t, err)

	_, err = io.Copy(part, bytes.NewReader(content))
	assert.NoError(t, err)

	err = writer.Close()
	assert.NoError(t, err)

	return body, writer.FormDataContentType()
}

func TestEpisodeHandler_UploadAudio(t *testing.T) {
	userID := uuid.New().String()
	channelID := uuid.New().String()
	episodeID := uuid.New().String()

	t.Run("音声ファイルをアップロードできる", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		episodeResp := createTestEpisodeResponse()
		result := &response.EpisodeDataResponse{Data: episodeResp}
		mockSvc.On("UploadAudio", mock.Anything, userID, channelID, episodeID, mock.AnythingOfType("service.UploadAudioInput")).Return(result, nil)

		handler := NewEpisodeHandler(mockSvc)
		router := setupAuthenticatedEpisodeRouter(handler, userID)

		body, contentType := createEpisodeAudioMultipartForm(t, "test.mp3", []byte("fake audio data"))
		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/channels/"+channelID+"/episodes/"+episodeID+"/audio", body)
		req.Header.Set("Content-Type", contentType)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp response.EpisodeDataResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.Data.ID)
		mockSvc.AssertExpectations(t)
	})

	t.Run("ファイルが指定されていない場合は 400 を返す", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		handler := NewEpisodeHandler(mockSvc)
		router := setupAuthenticatedEpisodeRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/channels/"+channelID+"/episodes/"+episodeID+"/audio", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("権限がない場合は 403 を返す", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		mockSvc.On("UploadAudio", mock.Anything, userID, channelID, episodeID, mock.AnythingOfType("service.UploadAudioInput")).Return(nil, apperror.ErrForbidden.WithMessage("このエピソードの音声アップロード権限がありません"))

		handler := NewEpisodeHandler(mockSvc)
		router := setupAuthenticatedEpisodeRouter(handler, userID)

		body, contentType := createEpisodeAudioMultipartForm(t, "test.mp3", []byte("fake audio data"))
		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/channels/"+channelID+"/episodes/"+episodeID+"/audio", body)
		req.Header.Set("Content-Type", contentType)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("エピソードが見つからない場合は 404 を返す", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		mockSvc.On("UploadAudio", mock.Anything, userID, channelID, episodeID, mock.AnythingOfType("service.UploadAudioInput")).Return(nil, apperror.ErrNotFound.WithMessage("エピソードが見つかりません"))

		handler := NewEpisodeHandler(mockSvc)
		router := setupAuthenticatedEpisodeRouter(handler, userID)

		body, contentType := createEpisodeAudioMultipartForm(t, "test.mp3", []byte("fake audio data"))
		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/channels/"+channelID+"/episodes/"+episodeID+"/audio", body)
		req.Header.Set("Content-Type", contentType)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("サービスがエラーを返すとエラーレスポンスを返す", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		mockSvc.On("UploadAudio", mock.Anything, userID, channelID, episodeID, mock.AnythingOfType("service.UploadAudioInput")).Return(nil, apperror.ErrInternal)

		handler := NewEpisodeHandler(mockSvc)
		router := setupAuthenticatedEpisodeRouter(handler, userID)

		body, contentType := createEpisodeAudioMultipartForm(t, "test.mp3", []byte("fake audio data"))
		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/channels/"+channelID+"/episodes/"+episodeID+"/audio", body)
		req.Header.Set("Content-Type", contentType)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("未認証の場合は 401 を返す", func(t *testing.T) {
		mockSvc := new(mockEpisodeService)
		handler := NewEpisodeHandler(mockSvc)
		router := setupEpisodeRouter(handler)

		body, contentType := createEpisodeAudioMultipartForm(t, "test.mp3", []byte("fake audio data"))
		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/channels/"+channelID+"/episodes/"+episodeID+"/audio", body)
		req.Header.Set("Content-Type", contentType)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
