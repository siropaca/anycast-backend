package handler

import (
	"bytes"
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

func (m *mockReactionService) GetReactionStatus(ctx context.Context, userID, episodeID string) (*response.ReactionStatusDataResponse, error) {
	args := m.Called(ctx, userID, episodeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.ReactionStatusDataResponse), args.Error(1)
}

func (m *mockReactionService) CreateOrUpdateReaction(ctx context.Context, userID, episodeID, reactionType string) (*response.ReactionDataResponse, bool, error) {
	args := m.Called(ctx, userID, episodeID, reactionType)
	if args.Get(0) == nil {
		return nil, false, args.Error(2)
	}
	return args.Get(0).(*response.ReactionDataResponse), args.Bool(1), args.Error(2)
}

func (m *mockReactionService) DeleteReaction(ctx context.Context, userID, episodeID string) error {
	args := m.Called(ctx, userID, episodeID)
	return args.Error(0)
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
	r.GET("/episodes/:episodeId/reactions", h.GetReactionStatus)
	r.POST("/episodes/:episodeId/reactions", h.CreateOrUpdateReaction)
	r.DELETE("/episodes/:episodeId/reactions", h.DeleteReaction)
	return r
}

// 認証なしのルーターをセットアップする
func setupReactionRouterWithoutAuth(h *ReactionHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/me/likes", h.ListLikes)
	r.GET("/episodes/:episodeId/reactions", h.GetReactionStatus)
	r.POST("/episodes/:episodeId/reactions", h.CreateOrUpdateReaction)
	r.DELETE("/episodes/:episodeId/reactions", h.DeleteReaction)
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

func TestReactionHandler_GetReactionStatus(t *testing.T) {
	userID := uuid.New().String()
	episodeID := uuid.New().String()

	t.Run("リアクション済みの場合 reactionType を返す", func(t *testing.T) {
		mockSvc := new(mockReactionService)
		reactionType := "like"
		result := &response.ReactionStatusDataResponse{
			Data: response.ReactionStatusResponse{
				ReactionType: &reactionType,
			},
		}
		mockSvc.On("GetReactionStatus", mock.Anything, userID, episodeID).Return(result, nil)

		handler := NewReactionHandler(mockSvc)
		router := setupReactionRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/episodes/"+episodeID+"/reactions", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp response.ReactionStatusDataResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.NotNil(t, resp.Data.ReactionType)
		assert.Equal(t, "like", *resp.Data.ReactionType)
		mockSvc.AssertExpectations(t)
	})

	t.Run("未リアクションの場合 reactionType が null を返す", func(t *testing.T) {
		mockSvc := new(mockReactionService)
		result := &response.ReactionStatusDataResponse{
			Data: response.ReactionStatusResponse{
				ReactionType: nil,
			},
		}
		mockSvc.On("GetReactionStatus", mock.Anything, userID, episodeID).Return(result, nil)

		handler := NewReactionHandler(mockSvc)
		router := setupReactionRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/episodes/"+episodeID+"/reactions", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp response.ReactionStatusDataResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Nil(t, resp.Data.ReactionType)
		mockSvc.AssertExpectations(t)
	})

	t.Run("認証されていない場合は 401 を返す", func(t *testing.T) {
		mockSvc := new(mockReactionService)

		handler := NewReactionHandler(mockSvc)
		router := setupReactionRouterWithoutAuth(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/episodes/"+episodeID+"/reactions", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockSvc.AssertNotCalled(t, "GetReactionStatus")
	})

	t.Run("サービスがエラーを返すとエラーレスポンスを返す", func(t *testing.T) {
		mockSvc := new(mockReactionService)
		mockSvc.On("GetReactionStatus", mock.Anything, userID, episodeID).Return(nil, apperror.ErrInternal)

		handler := NewReactionHandler(mockSvc)
		router := setupReactionRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/episodes/"+episodeID+"/reactions", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})
}

func TestReactionHandler_CreateOrUpdateReaction(t *testing.T) {
	userID := uuid.New().String()
	episodeID := uuid.New().String()

	t.Run("新規リアクション作成時に 201 を返す", func(t *testing.T) {
		mockSvc := new(mockReactionService)
		now := time.Now()
		result := &response.ReactionDataResponse{
			Data: response.ReactionResponse{
				ID:           uuid.New(),
				EpisodeID:    uuid.MustParse(episodeID),
				ReactionType: "like",
				CreatedAt:    now,
			},
		}
		mockSvc.On("CreateOrUpdateReaction", mock.Anything, userID, episodeID, "like").Return(result, true, nil)

		handler := NewReactionHandler(mockSvc)
		router := setupReactionRouter(handler, userID)

		body := bytes.NewBufferString(`{"reactionType":"like"}`)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/episodes/"+episodeID+"/reactions", body)
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp response.ReactionDataResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "like", resp.Data.ReactionType)
		mockSvc.AssertExpectations(t)
	})

	t.Run("既存リアクション更新時に 200 を返す", func(t *testing.T) {
		mockSvc := new(mockReactionService)
		now := time.Now()
		result := &response.ReactionDataResponse{
			Data: response.ReactionResponse{
				ID:           uuid.New(),
				EpisodeID:    uuid.MustParse(episodeID),
				ReactionType: "bad",
				CreatedAt:    now,
			},
		}
		mockSvc.On("CreateOrUpdateReaction", mock.Anything, userID, episodeID, "bad").Return(result, false, nil)

		handler := NewReactionHandler(mockSvc)
		router := setupReactionRouter(handler, userID)

		body := bytes.NewBufferString(`{"reactionType":"bad"}`)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/episodes/"+episodeID+"/reactions", body)
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp response.ReactionDataResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "bad", resp.Data.ReactionType)
		mockSvc.AssertExpectations(t)
	})

	t.Run("バリデーションエラーの場合は 400 を返す", func(t *testing.T) {
		mockSvc := new(mockReactionService)

		handler := NewReactionHandler(mockSvc)
		router := setupReactionRouter(handler, userID)

		body := bytes.NewBufferString(`{"reactionType":"invalid"}`)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/episodes/"+episodeID+"/reactions", body)
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockSvc.AssertNotCalled(t, "CreateOrUpdateReaction")
	})

	t.Run("リクエストボディが空の場合は 400 を返す", func(t *testing.T) {
		mockSvc := new(mockReactionService)

		handler := NewReactionHandler(mockSvc)
		router := setupReactionRouter(handler, userID)

		body := bytes.NewBufferString(`{}`)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/episodes/"+episodeID+"/reactions", body)
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockSvc.AssertNotCalled(t, "CreateOrUpdateReaction")
	})

	t.Run("認証されていない場合は 401 を返す", func(t *testing.T) {
		mockSvc := new(mockReactionService)

		handler := NewReactionHandler(mockSvc)
		router := setupReactionRouterWithoutAuth(handler)

		body := bytes.NewBufferString(`{"reactionType":"like"}`)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/episodes/"+episodeID+"/reactions", body)
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockSvc.AssertNotCalled(t, "CreateOrUpdateReaction")
	})

	t.Run("サービスがエラーを返すとエラーレスポンスを返す", func(t *testing.T) {
		mockSvc := new(mockReactionService)
		mockSvc.On("CreateOrUpdateReaction", mock.Anything, userID, episodeID, "like").Return(nil, false, apperror.ErrInternal)

		handler := NewReactionHandler(mockSvc)
		router := setupReactionRouter(handler, userID)

		body := bytes.NewBufferString(`{"reactionType":"like"}`)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/episodes/"+episodeID+"/reactions", body)
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})
}

func TestReactionHandler_DeleteReaction(t *testing.T) {
	userID := uuid.New().String()
	episodeID := uuid.New().String()

	t.Run("リアクション削除時に 204 を返す", func(t *testing.T) {
		mockSvc := new(mockReactionService)
		mockSvc.On("DeleteReaction", mock.Anything, userID, episodeID).Return(nil)

		handler := NewReactionHandler(mockSvc)
		router := setupReactionRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/episodes/"+episodeID+"/reactions", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("認証されていない場合は 401 を返す", func(t *testing.T) {
		mockSvc := new(mockReactionService)

		handler := NewReactionHandler(mockSvc)
		router := setupReactionRouterWithoutAuth(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/episodes/"+episodeID+"/reactions", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockSvc.AssertNotCalled(t, "DeleteReaction")
	})

	t.Run("存在しないリアクション削除時に 404 を返す", func(t *testing.T) {
		mockSvc := new(mockReactionService)
		mockSvc.On("DeleteReaction", mock.Anything, userID, episodeID).Return(apperror.ErrNotFound.WithMessage("リアクションが見つかりません"))

		handler := NewReactionHandler(mockSvc)
		router := setupReactionRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/episodes/"+episodeID+"/reactions", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("サービスがエラーを返すとエラーレスポンスを返す", func(t *testing.T) {
		mockSvc := new(mockReactionService)
		mockSvc.On("DeleteReaction", mock.Anything, userID, episodeID).Return(apperror.ErrInternal)

		handler := NewReactionHandler(mockSvc)
		router := setupReactionRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/episodes/"+episodeID+"/reactions", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})
}
