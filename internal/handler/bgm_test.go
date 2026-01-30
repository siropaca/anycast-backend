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
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/middleware"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// BgmService のモック
type mockBgmService struct {
	mock.Mock
}

func (m *mockBgmService) ListMyBgms(ctx context.Context, userID string, req request.ListMyBgmsRequest) (*response.BgmListWithPaginationResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.BgmListWithPaginationResponse), args.Error(1)
}

func (m *mockBgmService) GetMyBgm(ctx context.Context, userID, bgmID string) (*response.BgmDataResponse, error) {
	args := m.Called(ctx, userID, bgmID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.BgmDataResponse), args.Error(1)
}

func (m *mockBgmService) CreateBgm(ctx context.Context, userID string, req request.CreateBgmRequest) (*response.BgmDataResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.BgmDataResponse), args.Error(1)
}

func (m *mockBgmService) UpdateMyBgm(ctx context.Context, userID, bgmID string, req request.UpdateBgmRequest) (*response.BgmDataResponse, error) {
	args := m.Called(ctx, userID, bgmID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.BgmDataResponse), args.Error(1)
}

func (m *mockBgmService) DeleteMyBgm(ctx context.Context, userID, bgmID string) error {
	args := m.Called(ctx, userID, bgmID)
	return args.Error(0)
}

// ユーザー ID をコンテキストに設定するミドルウェア
func withBgmUserID(userID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(string(middleware.UserIDKey), userID)
		c.Next()
	}
}

// テスト用のルーターをセットアップする
func setupBgmRouter(h *BgmHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/me/bgms", h.ListMyBgms)
	r.POST("/me/bgms", h.CreateBgm)
	r.GET("/me/bgms/:bgmId", h.GetMyBgm)
	r.PATCH("/me/bgms/:bgmId", h.UpdateMyBgm)
	r.DELETE("/me/bgms/:bgmId", h.DeleteMyBgm)
	return r
}

// 認証済みルーターをセットアップする
func setupAuthenticatedBgmRouter(h *BgmHandler, userID string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(withBgmUserID(userID))
	r.GET("/me/bgms", h.ListMyBgms)
	r.POST("/me/bgms", h.CreateBgm)
	r.GET("/me/bgms/:bgmId", h.GetMyBgm)
	r.PATCH("/me/bgms/:bgmId", h.UpdateMyBgm)
	r.DELETE("/me/bgms/:bgmId", h.DeleteMyBgm)
	return r
}

// テスト用の BGM レスポンスを生成する
func createTestBgmWithEpisodesResponse() response.BgmWithEpisodesResponse {
	now := time.Now()
	return response.BgmWithEpisodesResponse{
		ID:       uuid.New(),
		Name:     "Test BGM",
		IsSystem: false,
		Audio: response.BgmAudioResponse{
			ID:         uuid.New(),
			URL:        "https://example.com/audio.mp3",
			DurationMs: 180000,
		},
		Episodes:  []response.BgmEpisodeResponse{},
		Channels:  []response.BgmChannelResponse{},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestBgmHandler_GetMyBgm(t *testing.T) {
	userID := uuid.New().String()
	bgmID := uuid.New().String()

	t.Run("自分の BGM を取得できる", func(t *testing.T) {
		mockSvc := new(mockBgmService)
		bgmResp := createTestBgmWithEpisodesResponse()
		result := &response.BgmDataResponse{Data: bgmResp}
		mockSvc.On("GetMyBgm", mock.Anything, userID, bgmID).Return(result, nil)

		handler := NewBgmHandler(mockSvc)
		router := setupAuthenticatedBgmRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me/bgms/"+bgmID, http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp response.BgmDataResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.Data.ID)
		mockSvc.AssertExpectations(t)
	})

	t.Run("BGM が見つからない場合は 404 を返す", func(t *testing.T) {
		mockSvc := new(mockBgmService)
		mockSvc.On("GetMyBgm", mock.Anything, userID, bgmID).Return(nil, apperror.ErrNotFound.WithMessage("BGM が見つかりません"))

		handler := NewBgmHandler(mockSvc)
		router := setupAuthenticatedBgmRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me/bgms/"+bgmID, http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("サービスがエラーを返すとエラーレスポンスを返す", func(t *testing.T) {
		mockSvc := new(mockBgmService)
		mockSvc.On("GetMyBgm", mock.Anything, userID, bgmID).Return(nil, apperror.ErrInternal)

		handler := NewBgmHandler(mockSvc)
		router := setupAuthenticatedBgmRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me/bgms/"+bgmID, http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("未認証の場合は 401 を返す", func(t *testing.T) {
		mockSvc := new(mockBgmService)
		handler := NewBgmHandler(mockSvc)
		router := setupBgmRouter(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me/bgms/"+bgmID, http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestBgmHandler_UpdateMyBgm(t *testing.T) {
	userID := uuid.New().String()
	bgmID := uuid.New().String()

	t.Run("BGM を更新できる", func(t *testing.T) {
		mockSvc := new(mockBgmService)
		bgmResp := createTestBgmWithEpisodesResponse()
		bgmResp.Name = "Updated BGM"
		result := &response.BgmDataResponse{Data: bgmResp}
		mockSvc.On("UpdateMyBgm", mock.Anything, userID, bgmID, mock.AnythingOfType("request.UpdateBgmRequest")).Return(result, nil)

		handler := NewBgmHandler(mockSvc)
		router := setupAuthenticatedBgmRouter(handler, userID)

		newName := "Updated BGM"
		reqBody := request.UpdateBgmRequest{
			Name: &newName,
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("PATCH", "/me/bgms/"+bgmID, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp response.BgmDataResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "Updated BGM", resp.Data.Name)
		mockSvc.AssertExpectations(t)
	})

	t.Run("BGM が見つからない場合は 404 を返す", func(t *testing.T) {
		mockSvc := new(mockBgmService)
		mockSvc.On("UpdateMyBgm", mock.Anything, userID, bgmID, mock.AnythingOfType("request.UpdateBgmRequest")).Return(nil, apperror.ErrNotFound.WithMessage("BGM が見つかりません"))

		handler := NewBgmHandler(mockSvc)
		router := setupAuthenticatedBgmRouter(handler, userID)

		newName := "Updated BGM"
		reqBody := request.UpdateBgmRequest{
			Name: &newName,
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("PATCH", "/me/bgms/"+bgmID, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("重複する名前の場合は 409 を返す", func(t *testing.T) {
		mockSvc := new(mockBgmService)
		mockSvc.On("UpdateMyBgm", mock.Anything, userID, bgmID, mock.AnythingOfType("request.UpdateBgmRequest")).Return(nil, apperror.ErrDuplicateName.WithMessage("同じ名前の BGM が既に存在します"))

		handler := NewBgmHandler(mockSvc)
		router := setupAuthenticatedBgmRouter(handler, userID)

		newName := "Duplicate Name"
		reqBody := request.UpdateBgmRequest{
			Name: &newName,
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("PATCH", "/me/bgms/"+bgmID, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("未認証の場合は 401 を返す", func(t *testing.T) {
		mockSvc := new(mockBgmService)
		handler := NewBgmHandler(mockSvc)
		router := setupBgmRouter(handler)

		newName := "Updated BGM"
		reqBody := request.UpdateBgmRequest{
			Name: &newName,
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("PATCH", "/me/bgms/"+bgmID, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestBgmHandler_DeleteMyBgm(t *testing.T) {
	userID := uuid.New().String()
	bgmID := uuid.New().String()

	t.Run("BGM を削除できる", func(t *testing.T) {
		mockSvc := new(mockBgmService)
		mockSvc.On("DeleteMyBgm", mock.Anything, userID, bgmID).Return(nil)

		handler := NewBgmHandler(mockSvc)
		router := setupAuthenticatedBgmRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/me/bgms/"+bgmID, http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Empty(t, w.Body.String())
		mockSvc.AssertExpectations(t)
	})

	t.Run("BGM が見つからない場合は 404 を返す", func(t *testing.T) {
		mockSvc := new(mockBgmService)
		mockSvc.On("DeleteMyBgm", mock.Anything, userID, bgmID).Return(apperror.ErrNotFound.WithMessage("BGM が見つかりません"))

		handler := NewBgmHandler(mockSvc)
		router := setupAuthenticatedBgmRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/me/bgms/"+bgmID, http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("使用中の BGM は削除できない", func(t *testing.T) {
		mockSvc := new(mockBgmService)
		mockSvc.On("DeleteMyBgm", mock.Anything, userID, bgmID).Return(apperror.ErrBgmInUse.WithMessage("この BGM はエピソードで使用中のため削除できません"))

		handler := NewBgmHandler(mockSvc)
		router := setupAuthenticatedBgmRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/me/bgms/"+bgmID, http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("未認証の場合は 401 を返す", func(t *testing.T) {
		mockSvc := new(mockBgmService)
		handler := NewBgmHandler(mockSvc)
		router := setupBgmRouter(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/me/bgms/"+bgmID, http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
