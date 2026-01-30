package handler

import (
	"context"
	"encoding/json"
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
)

// RecommendationService のモック
type mockRecommendationService struct {
	mock.Mock
}

func (m *mockRecommendationService) GetRecommendedChannels(ctx context.Context, userID *string, req request.RecommendChannelsRequest) (*response.RecommendedChannelListResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.RecommendedChannelListResponse), args.Error(1)
}

// 未ログインのルーターをセットアップする
func setupRecommendationRouter(h *RecommendationHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/recommendations/channels", h.GetRecommendedChannels)
	return r
}

// ログイン済みのルーターをセットアップする
func setupRecommendationRouterWithAuth(h *RecommendationHandler, userID string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(string(middleware.UserIDKey), userID)
		c.Next()
	})
	r.GET("/recommendations/channels", h.GetRecommendedChannels)
	return r
}

func TestNewRecommendationHandler(t *testing.T) {
	t.Run("RecommendationHandler を作成できる", func(t *testing.T) {
		mockSvc := new(mockRecommendationService)
		handler := NewRecommendationHandler(mockSvc)
		assert.NotNil(t, handler)
	})
}

func TestRecommendationHandler_GetRecommendedChannels(t *testing.T) {
	now := time.Now()
	channelID := uuid.New()
	categoryID := uuid.New()

	baseResult := &response.RecommendedChannelListResponse{
		Data: []response.RecommendedChannelResponse{
			{
				ID:          channelID,
				Name:        "テストチャンネル",
				Description: "テスト説明",
				Category: response.CategoryResponse{
					ID:   categoryID,
					Slug: "technology",
					Name: "テクノロジー",
				},
				EpisodeCount:    5,
				TotalPlayCount:  100,
				LatestEpisodeAt: &now,
			},
		},
		Pagination: response.PaginationResponse{
			Total:  1,
			Limit:  20,
			Offset: 0,
		},
	}

	t.Run("未ログインでおすすめチャンネル一覧を取得できる", func(t *testing.T) {
		mockSvc := new(mockRecommendationService)
		mockSvc.On("GetRecommendedChannels", mock.Anything, (*string)(nil), mock.AnythingOfType("request.RecommendChannelsRequest")).Return(baseResult, nil)

		handler := NewRecommendationHandler(mockSvc)
		router := setupRecommendationRouter(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/recommendations/channels", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		data := resp["data"].([]any)
		assert.Len(t, data, 1)

		mockSvc.AssertExpectations(t)
	})

	t.Run("ログイン済みでおすすめチャンネル一覧を取得できる", func(t *testing.T) {
		mockSvc := new(mockRecommendationService)
		userID := uuid.New().String()
		mockSvc.On("GetRecommendedChannels", mock.Anything, &userID, mock.AnythingOfType("request.RecommendChannelsRequest")).Return(baseResult, nil)

		handler := NewRecommendationHandler(mockSvc)
		router := setupRecommendationRouterWithAuth(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/recommendations/channels", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("ページネーションパラメータを指定できる", func(t *testing.T) {
		mockSvc := new(mockRecommendationService)
		mockSvc.On("GetRecommendedChannels", mock.Anything, (*string)(nil), mock.AnythingOfType("request.RecommendChannelsRequest")).Return(baseResult, nil)

		handler := NewRecommendationHandler(mockSvc)
		router := setupRecommendationRouter(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/recommendations/channels?limit=5&offset=10", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("カテゴリ ID でフィルタできる", func(t *testing.T) {
		mockSvc := new(mockRecommendationService)
		mockSvc.On("GetRecommendedChannels", mock.Anything, (*string)(nil), mock.AnythingOfType("request.RecommendChannelsRequest")).Return(baseResult, nil)

		handler := NewRecommendationHandler(mockSvc)
		router := setupRecommendationRouter(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/recommendations/channels?categoryId="+categoryID.String(), http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("無効なカテゴリ ID の場合は 400 を返す", func(t *testing.T) {
		mockSvc := new(mockRecommendationService)

		handler := NewRecommendationHandler(mockSvc)
		router := setupRecommendationRouter(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/recommendations/channels?categoryId=invalid-uuid", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockSvc.AssertNotCalled(t, "GetRecommendedChannels")
	})

	t.Run("サービスがエラーを返すとエラーレスポンスを返す", func(t *testing.T) {
		mockSvc := new(mockRecommendationService)
		mockSvc.On("GetRecommendedChannels", mock.Anything, (*string)(nil), mock.AnythingOfType("request.RecommendChannelsRequest")).Return(nil, apperror.ErrInternal)

		handler := NewRecommendationHandler(mockSvc)
		router := setupRecommendationRouter(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/recommendations/channels", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("空のおすすめ一覧を取得できる", func(t *testing.T) {
		mockSvc := new(mockRecommendationService)
		emptyResult := &response.RecommendedChannelListResponse{
			Data: []response.RecommendedChannelResponse{},
			Pagination: response.PaginationResponse{
				Total:  0,
				Limit:  20,
				Offset: 0,
			},
		}
		mockSvc.On("GetRecommendedChannels", mock.Anything, (*string)(nil), mock.AnythingOfType("request.RecommendChannelsRequest")).Return(emptyResult, nil)

		handler := NewRecommendationHandler(mockSvc)
		router := setupRecommendationRouter(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/recommendations/channels", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		data := resp["data"].([]any)
		assert.Empty(t, data)

		mockSvc.AssertExpectations(t)
	})
}
