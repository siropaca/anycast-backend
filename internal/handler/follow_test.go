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

// FollowService のモック
type mockFollowService struct {
	mock.Mock
}

func (m *mockFollowService) ListFollows(ctx context.Context, userID string, limit, offset int) (*response.FollowListWithPaginationResponse, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.FollowListWithPaginationResponse), args.Error(1)
}

func (m *mockFollowService) CreateFollow(ctx context.Context, userID, targetUserID string) (*response.FollowDataResponse, error) {
	args := m.Called(ctx, userID, targetUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.FollowDataResponse), args.Error(1)
}

func (m *mockFollowService) DeleteFollow(ctx context.Context, userID, targetUserID string) error {
	args := m.Called(ctx, userID, targetUserID)
	return args.Error(0)
}

// テスト用のルーターをセットアップする
func setupFollowRouter(h *FollowHandler, userID string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(string(middleware.UserIDKey), userID)
		c.Next()
	})
	r.GET("/me/follows", h.ListFollows)
	r.POST("/users/:userId/follow", h.CreateFollow)
	r.DELETE("/users/:userId/follow", h.DeleteFollow)
	return r
}

// 認証なしのルーターをセットアップする
func setupFollowRouterWithoutAuth(h *FollowHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/me/follows", h.ListFollows)
	r.POST("/users/:userId/follow", h.CreateFollow)
	r.DELETE("/users/:userId/follow", h.DeleteFollow)
	return r
}

func TestNewFollowHandler(t *testing.T) {
	t.Run("FollowHandler を作成できる", func(t *testing.T) {
		mockSvc := new(mockFollowService)
		handler := NewFollowHandler(mockSvc)
		assert.NotNil(t, handler)
	})
}

func TestFollowHandler_ListFollows(t *testing.T) {
	userID := uuid.New().String()

	t.Run("フォロー中のユーザー一覧を取得できる", func(t *testing.T) {
		mockSvc := new(mockFollowService)
		now := time.Now()
		result := &response.FollowListWithPaginationResponse{
			Data: []response.FollowItemResponse{
				{
					User: response.FollowUserResponse{
						ID:          uuid.New(),
						Username:    "user1",
						DisplayName: "User 1",
					},
					FollowedAt: now,
				},
			},
			Pagination: response.PaginationResponse{
				Total:  1,
				Limit:  20,
				Offset: 0,
			},
		}
		mockSvc.On("ListFollows", mock.Anything, userID, 20, 0).Return(result, nil)

		handler := NewFollowHandler(mockSvc)
		router := setupFollowRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me/follows", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp response.FollowListWithPaginationResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Len(t, resp.Data, 1)
		assert.Equal(t, "user1", resp.Data[0].User.Username)
		assert.Equal(t, int64(1), resp.Pagination.Total)
		mockSvc.AssertExpectations(t)
	})

	t.Run("ページネーションパラメータを指定できる", func(t *testing.T) {
		mockSvc := new(mockFollowService)
		result := &response.FollowListWithPaginationResponse{
			Data: []response.FollowItemResponse{},
			Pagination: response.PaginationResponse{
				Total:  0,
				Limit:  10,
				Offset: 5,
			},
		}
		mockSvc.On("ListFollows", mock.Anything, userID, 10, 5).Return(result, nil)

		handler := NewFollowHandler(mockSvc)
		router := setupFollowRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me/follows?limit=10&offset=5", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("認証されていない場合は 401 を返す", func(t *testing.T) {
		mockSvc := new(mockFollowService)

		handler := NewFollowHandler(mockSvc)
		router := setupFollowRouterWithoutAuth(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me/follows", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockSvc.AssertNotCalled(t, "ListFollows")
	})

	t.Run("サービスがエラーを返すとエラーレスポンスを返す", func(t *testing.T) {
		mockSvc := new(mockFollowService)
		mockSvc.On("ListFollows", mock.Anything, userID, 20, 0).Return(nil, apperror.ErrInternal)

		handler := NewFollowHandler(mockSvc)
		router := setupFollowRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me/follows", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})
}

func TestFollowHandler_CreateFollow(t *testing.T) {
	userID := uuid.New().String()
	targetUserID := uuid.New().String()

	t.Run("フォロー登録時に 201 を返す", func(t *testing.T) {
		mockSvc := new(mockFollowService)
		now := time.Now()
		result := &response.FollowDataResponse{
			Data: response.FollowResponse{
				ID:           uuid.New(),
				TargetUserID: uuid.MustParse(targetUserID),
				CreatedAt:    now,
			},
		}
		mockSvc.On("CreateFollow", mock.Anything, userID, targetUserID).Return(result, nil)

		handler := NewFollowHandler(mockSvc)
		router := setupFollowRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/users/"+targetUserID+"/follow", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp response.FollowDataResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, uuid.MustParse(targetUserID), resp.Data.TargetUserID)
		mockSvc.AssertExpectations(t)
	})

	t.Run("自分自身のフォローは 400 を返す", func(t *testing.T) {
		mockSvc := new(mockFollowService)
		mockSvc.On("CreateFollow", mock.Anything, userID, userID).Return(nil, apperror.ErrSelfFollowNotAllowed)

		handler := NewFollowHandler(mockSvc)
		router := setupFollowRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/users/"+userID+"/follow", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("既にフォロー済みの場合は 409 を返す", func(t *testing.T) {
		mockSvc := new(mockFollowService)
		mockSvc.On("CreateFollow", mock.Anything, userID, targetUserID).Return(nil, apperror.ErrAlreadyFollowed)

		handler := NewFollowHandler(mockSvc)
		router := setupFollowRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/users/"+targetUserID+"/follow", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("認証されていない場合は 401 を返す", func(t *testing.T) {
		mockSvc := new(mockFollowService)

		handler := NewFollowHandler(mockSvc)
		router := setupFollowRouterWithoutAuth(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/users/"+targetUserID+"/follow", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockSvc.AssertNotCalled(t, "CreateFollow")
	})

	t.Run("サービスがエラーを返すとエラーレスポンスを返す", func(t *testing.T) {
		mockSvc := new(mockFollowService)
		mockSvc.On("CreateFollow", mock.Anything, userID, targetUserID).Return(nil, apperror.ErrInternal)

		handler := NewFollowHandler(mockSvc)
		router := setupFollowRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/users/"+targetUserID+"/follow", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})
}

func TestFollowHandler_DeleteFollow(t *testing.T) {
	userID := uuid.New().String()
	targetUserID := uuid.New().String()

	t.Run("フォロー解除時に 204 を返す", func(t *testing.T) {
		mockSvc := new(mockFollowService)
		mockSvc.On("DeleteFollow", mock.Anything, userID, targetUserID).Return(nil)

		handler := NewFollowHandler(mockSvc)
		router := setupFollowRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/users/"+targetUserID+"/follow", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("認証されていない場合は 401 を返す", func(t *testing.T) {
		mockSvc := new(mockFollowService)

		handler := NewFollowHandler(mockSvc)
		router := setupFollowRouterWithoutAuth(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/users/"+targetUserID+"/follow", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockSvc.AssertNotCalled(t, "DeleteFollow")
	})

	t.Run("存在しないフォロー解除時に 404 を返す", func(t *testing.T) {
		mockSvc := new(mockFollowService)
		mockSvc.On("DeleteFollow", mock.Anything, userID, targetUserID).Return(apperror.ErrNotFound.WithMessage("フォローが見つかりません"))

		handler := NewFollowHandler(mockSvc)
		router := setupFollowRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/users/"+targetUserID+"/follow", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("サービスがエラーを返すとエラーレスポンスを返す", func(t *testing.T) {
		mockSvc := new(mockFollowService)
		mockSvc.On("DeleteFollow", mock.Anything, userID, targetUserID).Return(apperror.ErrInternal)

		handler := NewFollowHandler(mockSvc)
		router := setupFollowRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/users/"+targetUserID+"/follow", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})
}
