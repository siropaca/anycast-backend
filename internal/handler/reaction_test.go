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
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/middleware"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// ReactionService のモック
type mockReactionService struct {
	mock.Mock
}

func (m *mockReactionService) ListLikes(ctx context.Context, userID string, limit, offset int) (*response.LikeListWithPaginationResponse, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.LikeListWithPaginationResponse), args.Error(1)
}

// テスト用のルーターをセットアップする
func setupReactionRouter(h *ReactionHandler, userID string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(string(middleware.UserIDKey), userID)
		c.Next()
	})
	r.GET("/me/likes", h.ListLikes)
	return r
}

// 認証なしのルーターをセットアップする
func setupReactionRouterWithoutAuth(h *ReactionHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/me/likes", h.ListLikes)
	return r
}

func TestNewReactionHandler(t *testing.T) {
	t.Run("ReactionHandler を作成できる", func(t *testing.T) {
		mockSvc := new(mockReactionService)
		handler := NewReactionHandler(mockSvc)
		assert.NotNil(t, handler)
	})
}

func TestReactionHandler_ListLikes(t *testing.T) {
	userID := uuid.New().String()

	t.Run("高評価したエピソード一覧を取得できる", func(t *testing.T) {
		mockSvc := new(mockReactionService)
		now := time.Now()
		result := &response.LikeListWithPaginationResponse{
			Data: []response.LikeItemResponse{
				{
					Episode: response.LikeEpisodeResponse{
						ID:          uuid.New(),
						Title:       "テストエピソード",
						Description: "テスト説明",
						Channel: response.LikeChannelResponse{
							ID:   uuid.New(),
							Name: "テストチャンネル",
						},
						PublishedAt: &now,
					},
					LikedAt: now,
				},
			},
			Pagination: response.PaginationResponse{
				Total:  1,
				Limit:  20,
				Offset: 0,
			},
		}
		mockSvc.On("ListLikes", mock.Anything, userID, 20, 0).Return(result, nil)

		handler := NewReactionHandler(mockSvc)
		router := setupReactionRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me/likes", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp response.LikeListWithPaginationResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Len(t, resp.Data, 1)
		assert.Equal(t, "テストエピソード", resp.Data[0].Episode.Title)
		assert.Equal(t, int64(1), resp.Pagination.Total)
		mockSvc.AssertExpectations(t)
	})

	t.Run("ページネーションパラメータを指定できる", func(t *testing.T) {
		mockSvc := new(mockReactionService)
		result := &response.LikeListWithPaginationResponse{
			Data: []response.LikeItemResponse{},
			Pagination: response.PaginationResponse{
				Total:  0,
				Limit:  10,
				Offset: 5,
			},
		}
		mockSvc.On("ListLikes", mock.Anything, userID, 10, 5).Return(result, nil)

		handler := NewReactionHandler(mockSvc)
		router := setupReactionRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me/likes?limit=10&offset=5", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("認証されていない場合は 401 を返す", func(t *testing.T) {
		mockSvc := new(mockReactionService)

		handler := NewReactionHandler(mockSvc)
		router := setupReactionRouterWithoutAuth(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me/likes", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockSvc.AssertNotCalled(t, "ListLikes")
	})

	t.Run("サービスがエラーを返すとエラーレスポンスを返す", func(t *testing.T) {
		mockSvc := new(mockReactionService)
		mockSvc.On("ListLikes", mock.Anything, userID, 20, 0).Return(nil, apperror.ErrInternal)

		handler := NewReactionHandler(mockSvc)
		router := setupReactionRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me/likes", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})
}
