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
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/middleware"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// ChannelService のモック
type mockChannelService struct {
	mock.Mock
}

func (m *mockChannelService) GetChannel(ctx context.Context, userID, channelID string) (*response.ChannelDataResponse, error) {
	args := m.Called(ctx, userID, channelID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.ChannelDataResponse), args.Error(1)
}

func (m *mockChannelService) GetMyChannel(ctx context.Context, userID, channelID string) (*response.ChannelDataResponse, error) {
	args := m.Called(ctx, userID, channelID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.ChannelDataResponse), args.Error(1)
}

func (m *mockChannelService) ListMyChannels(ctx context.Context, userID string, filter repository.ChannelFilter) (*response.ChannelListWithPaginationResponse, error) {
	args := m.Called(ctx, userID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.ChannelListWithPaginationResponse), args.Error(1)
}

func (m *mockChannelService) CreateChannel(ctx context.Context, userID string, req request.CreateChannelRequest) (*response.ChannelDataResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.ChannelDataResponse), args.Error(1)
}

func (m *mockChannelService) UpdateChannel(ctx context.Context, userID, channelID string, req request.UpdateChannelRequest) (*response.ChannelDataResponse, error) {
	args := m.Called(ctx, userID, channelID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.ChannelDataResponse), args.Error(1)
}

func (m *mockChannelService) DeleteChannel(ctx context.Context, userID, channelID string) error {
	args := m.Called(ctx, userID, channelID)
	return args.Error(0)
}

func (m *mockChannelService) PublishChannel(ctx context.Context, userID, channelID string, publishedAt *string) (*response.ChannelDataResponse, error) {
	args := m.Called(ctx, userID, channelID, publishedAt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.ChannelDataResponse), args.Error(1)
}

func (m *mockChannelService) UnpublishChannel(ctx context.Context, userID, channelID string) (*response.ChannelDataResponse, error) {
	args := m.Called(ctx, userID, channelID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.ChannelDataResponse), args.Error(1)
}

// テスト用のルーターをセットアップする
func setupChannelRouter(h *ChannelHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/me/channels", h.ListMyChannels)
	r.GET("/me/channels/:channelId", h.GetMyChannel)
	r.POST("/channels", h.CreateChannel)
	r.GET("/channels/:channelId", h.GetChannel)
	r.PATCH("/channels/:channelId", h.UpdateChannel)
	r.DELETE("/channels/:channelId", h.DeleteChannel)
	return r
}

// ユーザー ID をコンテキストに設定するミドルウェア
func withUserID(userID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(string(middleware.UserIDKey), userID)
		c.Next()
	}
}

// 認証済みルーターをセットアップする
func setupAuthenticatedChannelRouter(h *ChannelHandler, userID string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(withUserID(userID))
	r.GET("/me/channels", h.ListMyChannels)
	r.GET("/me/channels/:channelId", h.GetMyChannel)
	r.POST("/channels", h.CreateChannel)
	r.GET("/channels/:channelId", h.GetChannel)
	r.PATCH("/channels/:channelId", h.UpdateChannel)
	r.DELETE("/channels/:channelId", h.DeleteChannel)
	return r
}

// テスト用のチャンネルレスポンスを生成する
func createTestChannelResponse() response.ChannelResponse {
	now := time.Now()
	return response.ChannelResponse{
		ID:          uuid.New(),
		Name:        "Test Channel",
		Description: "Test Description",
		UserPrompt:  "Test User Prompt",
		Category: response.CategoryResponse{
			ID:        uuid.New(),
			Slug:      "technology",
			Name:      "テクノロジー",
			SortOrder: 0,
			IsActive:  true,
		},
		Characters: []response.CharacterResponse{
			{
				ID:      uuid.New(),
				Name:    "Host",
				Persona: "Friendly host",
				Voice: response.CharacterVoiceResponse{
					ID:     uuid.New(),
					Name:   "Voice 1",
					Gender: "male",
				},
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestChannelHandler_ListMyChannels(t *testing.T) {
	userID := uuid.New().String()

	t.Run("チャンネル一覧を取得できる", func(t *testing.T) {
		mockSvc := new(mockChannelService)
		channels := []response.ChannelResponse{createTestChannelResponse()}
		result := &response.ChannelListWithPaginationResponse{
			Data:       channels,
			Pagination: response.PaginationResponse{Total: 1, Limit: 20, Offset: 0},
		}
		mockSvc.On("ListMyChannels", mock.Anything, userID, mock.AnythingOfType("repository.ChannelFilter")).Return(result, nil)

		handler := NewChannelHandler(mockSvc)
		router := setupAuthenticatedChannelRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me/channels", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp response.ChannelListWithPaginationResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Len(t, resp.Data, 1)
		mockSvc.AssertExpectations(t)
	})

	t.Run("空のチャンネル一覧を取得できる", func(t *testing.T) {
		mockSvc := new(mockChannelService)
		result := &response.ChannelListWithPaginationResponse{
			Data:       []response.ChannelResponse{},
			Pagination: response.PaginationResponse{Total: 0, Limit: 20, Offset: 0},
		}
		mockSvc.On("ListMyChannels", mock.Anything, userID, mock.AnythingOfType("repository.ChannelFilter")).Return(result, nil)

		handler := NewChannelHandler(mockSvc)
		router := setupAuthenticatedChannelRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me/channels", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp response.ChannelListWithPaginationResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Empty(t, resp.Data)
		mockSvc.AssertExpectations(t)
	})

	t.Run("クエリパラメータでフィルタできる", func(t *testing.T) {
		mockSvc := new(mockChannelService)
		result := &response.ChannelListWithPaginationResponse{
			Data:       []response.ChannelResponse{},
			Pagination: response.PaginationResponse{Total: 0, Limit: 10, Offset: 5},
		}
		mockSvc.On("ListMyChannels", mock.Anything, userID, mock.MatchedBy(func(f repository.ChannelFilter) bool {
			return f.Limit == 10 && f.Offset == 5 && *f.Status == "published"
		})).Return(result, nil)

		handler := NewChannelHandler(mockSvc)
		router := setupAuthenticatedChannelRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me/channels?status=published&limit=10&offset=5", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("サービスがエラーを返すとエラーレスポンスを返す", func(t *testing.T) {
		mockSvc := new(mockChannelService)
		mockSvc.On("ListMyChannels", mock.Anything, userID, mock.AnythingOfType("repository.ChannelFilter")).Return(nil, apperror.ErrInternal)

		handler := NewChannelHandler(mockSvc)
		router := setupAuthenticatedChannelRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me/channels", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("未認証の場合は 401 を返す", func(t *testing.T) {
		mockSvc := new(mockChannelService)
		handler := NewChannelHandler(mockSvc)
		router := setupChannelRouter(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me/channels", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestChannelHandler_GetMyChannel(t *testing.T) {
	userID := uuid.New().String()
	channelID := uuid.New().String()

	t.Run("自分のチャンネルを取得できる", func(t *testing.T) {
		mockSvc := new(mockChannelService)
		channelResp := createTestChannelResponse()
		result := &response.ChannelDataResponse{Data: channelResp}
		mockSvc.On("GetMyChannel", mock.Anything, userID, channelID).Return(result, nil)

		handler := NewChannelHandler(mockSvc)
		router := setupAuthenticatedChannelRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me/channels/"+channelID, http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp response.ChannelDataResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.Data.ID)
		mockSvc.AssertExpectations(t)
	})

	t.Run("チャンネルが見つからない場合は 404 を返す", func(t *testing.T) {
		mockSvc := new(mockChannelService)
		mockSvc.On("GetMyChannel", mock.Anything, userID, channelID).Return(nil, apperror.ErrNotFound.WithMessage("Channel not found"))

		handler := NewChannelHandler(mockSvc)
		router := setupAuthenticatedChannelRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me/channels/"+channelID, http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("権限がない場合は 403 を返す", func(t *testing.T) {
		mockSvc := new(mockChannelService)
		mockSvc.On("GetMyChannel", mock.Anything, userID, channelID).Return(nil, apperror.ErrForbidden.WithMessage("You do not have permission"))

		handler := NewChannelHandler(mockSvc)
		router := setupAuthenticatedChannelRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me/channels/"+channelID, http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("サービスがエラーを返すとエラーレスポンスを返す", func(t *testing.T) {
		mockSvc := new(mockChannelService)
		mockSvc.On("GetMyChannel", mock.Anything, userID, channelID).Return(nil, apperror.ErrInternal)

		handler := NewChannelHandler(mockSvc)
		router := setupAuthenticatedChannelRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me/channels/"+channelID, http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("未認証の場合は 401 を返す", func(t *testing.T) {
		mockSvc := new(mockChannelService)
		handler := NewChannelHandler(mockSvc)
		router := setupChannelRouter(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me/channels/"+channelID, http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestChannelHandler_CreateChannel(t *testing.T) {
	userID := uuid.New().String()
	categoryID := uuid.New().String()
	voiceID := uuid.New().String()

	t.Run("チャンネルを作成できる", func(t *testing.T) {
		mockSvc := new(mockChannelService)
		channelResp := createTestChannelResponse()
		result := &response.ChannelDataResponse{Data: channelResp}

		mockSvc.On("CreateChannel", mock.Anything, userID, mock.AnythingOfType("request.CreateChannelRequest")).Return(result, nil)

		handler := NewChannelHandler(mockSvc)
		router := setupAuthenticatedChannelRouter(handler, userID)

		hostName := "Host"
		hostPersona := "Friendly"
		reqBody := request.CreateChannelRequest{
			Name:        "New Channel",
			Description: "Description",
			UserPrompt:  "User prompt",
			CategoryID:  categoryID,
			Characters: []request.ChannelCharacterInputRequest{
				{Name: &hostName, Persona: &hostPersona, VoiceID: &voiceID},
			},
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/channels", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp response.ChannelDataResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.Data.ID)
		mockSvc.AssertExpectations(t)
	})

	t.Run("バリデーションエラーの場合は 400 を返す", func(t *testing.T) {
		mockSvc := new(mockChannelService)
		handler := NewChannelHandler(mockSvc)
		router := setupAuthenticatedChannelRouter(handler, userID)

		// 必須フィールドが欠けているリクエスト
		reqBody := map[string]any{
			"name": "Channel without required fields",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/channels", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("サービスがエラーを返すとエラーレスポンスを返す", func(t *testing.T) {
		mockSvc := new(mockChannelService)
		mockSvc.On("CreateChannel", mock.Anything, userID, mock.AnythingOfType("request.CreateChannelRequest")).Return(nil, apperror.ErrNotFound.WithMessage("Category not found"))

		handler := NewChannelHandler(mockSvc)
		router := setupAuthenticatedChannelRouter(handler, userID)

		hostName := "Host"
		hostPersona := "Friendly"
		reqBody := request.CreateChannelRequest{
			Name:        "New Channel",
			Description: "Description",
			UserPrompt:  "User prompt",
			CategoryID:  categoryID,
			Characters: []request.ChannelCharacterInputRequest{
				{Name: &hostName, Persona: &hostPersona, VoiceID: &voiceID},
			},
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/channels", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("未認証の場合は 401 を返す", func(t *testing.T) {
		mockSvc := new(mockChannelService)
		handler := NewChannelHandler(mockSvc)
		router := setupChannelRouter(handler)

		hostName := "Host"
		hostPersona := "Friendly"
		reqBody := request.CreateChannelRequest{
			Name:        "New Channel",
			Description: "Description",
			UserPrompt:  "User prompt",
			CategoryID:  categoryID,
			Characters: []request.ChannelCharacterInputRequest{
				{Name: &hostName, Persona: &hostPersona, VoiceID: &voiceID},
			},
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/channels", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestChannelHandler_GetChannel(t *testing.T) {
	userID := uuid.New().String()
	channelID := uuid.New().String()

	t.Run("チャンネルを取得できる", func(t *testing.T) {
		mockSvc := new(mockChannelService)
		channelResp := createTestChannelResponse()
		result := &response.ChannelDataResponse{Data: channelResp}
		mockSvc.On("GetChannel", mock.Anything, userID, channelID).Return(result, nil)

		handler := NewChannelHandler(mockSvc)
		router := setupAuthenticatedChannelRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/channels/"+channelID, http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp response.ChannelDataResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.Data.ID)
		mockSvc.AssertExpectations(t)
	})

	t.Run("チャンネルが見つからない場合は 404 を返す", func(t *testing.T) {
		mockSvc := new(mockChannelService)
		mockSvc.On("GetChannel", mock.Anything, userID, channelID).Return(nil, apperror.ErrNotFound.WithMessage("Channel not found"))

		handler := NewChannelHandler(mockSvc)
		router := setupAuthenticatedChannelRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/channels/"+channelID, http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("未認証の場合は 401 を返す", func(t *testing.T) {
		mockSvc := new(mockChannelService)
		handler := NewChannelHandler(mockSvc)
		router := setupChannelRouter(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/channels/"+channelID, http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestChannelHandler_UpdateChannel(t *testing.T) {
	userID := uuid.New().String()
	channelID := uuid.New().String()

	t.Run("チャンネルを更新できる", func(t *testing.T) {
		mockSvc := new(mockChannelService)
		channelResp := createTestChannelResponse()
		channelResp.Name = "Updated Channel"
		result := &response.ChannelDataResponse{Data: channelResp}
		mockSvc.On("UpdateChannel", mock.Anything, userID, channelID, mock.AnythingOfType("request.UpdateChannelRequest")).Return(result, nil)

		handler := NewChannelHandler(mockSvc)
		router := setupAuthenticatedChannelRouter(handler, userID)

		reqBody := request.UpdateChannelRequest{
			Name:        "Updated Channel",
			Description: "Updated Description",
			UserPrompt:  "Updated User Prompt",
			CategoryID:  uuid.New().String(),
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("PATCH", "/channels/"+channelID, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp response.ChannelDataResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Channel", resp.Data.Name)
		mockSvc.AssertExpectations(t)
	})

	t.Run("必須フィールドが欠けているとバリデーションエラーを返す", func(t *testing.T) {
		mockSvc := new(mockChannelService)
		handler := NewChannelHandler(mockSvc)
		router := setupAuthenticatedChannelRouter(handler, userID)

		// 必須フィールドが欠けているリクエスト
		reqBody := map[string]any{
			"name": "Channel without required fields",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("PATCH", "/channels/"+channelID, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("権限がない場合は 403 を返す", func(t *testing.T) {
		mockSvc := new(mockChannelService)
		mockSvc.On("UpdateChannel", mock.Anything, userID, channelID, mock.AnythingOfType("request.UpdateChannelRequest")).Return(nil, apperror.ErrForbidden.WithMessage("You do not have permission"))

		handler := NewChannelHandler(mockSvc)
		router := setupAuthenticatedChannelRouter(handler, userID)

		reqBody := request.UpdateChannelRequest{
			Name:        "Updated Channel",
			Description: "Updated Description",
			UserPrompt:  "Updated User Prompt",
			CategoryID:  uuid.New().String(),
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("PATCH", "/channels/"+channelID, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("チャンネルが見つからない場合は 404 を返す", func(t *testing.T) {
		mockSvc := new(mockChannelService)
		mockSvc.On("UpdateChannel", mock.Anything, userID, channelID, mock.AnythingOfType("request.UpdateChannelRequest")).Return(nil, apperror.ErrNotFound.WithMessage("Channel not found"))

		handler := NewChannelHandler(mockSvc)
		router := setupAuthenticatedChannelRouter(handler, userID)

		reqBody := request.UpdateChannelRequest{
			Name:        "Updated Channel",
			Description: "Updated Description",
			UserPrompt:  "Updated User Prompt",
			CategoryID:  uuid.New().String(),
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("PATCH", "/channels/"+channelID, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("未認証の場合は 401 を返す", func(t *testing.T) {
		mockSvc := new(mockChannelService)
		handler := NewChannelHandler(mockSvc)
		router := setupChannelRouter(handler)

		reqBody := request.UpdateChannelRequest{
			Name:        "Updated Channel",
			Description: "Updated Description",
			UserPrompt:  "Updated User Prompt",
			CategoryID:  uuid.New().String(),
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("PATCH", "/channels/"+channelID, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestChannelHandler_DeleteChannel(t *testing.T) {
	userID := uuid.New().String()
	channelID := uuid.New().String()

	t.Run("チャンネルを削除できる", func(t *testing.T) {
		mockSvc := new(mockChannelService)
		mockSvc.On("DeleteChannel", mock.Anything, userID, channelID).Return(nil)

		handler := NewChannelHandler(mockSvc)
		router := setupAuthenticatedChannelRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/channels/"+channelID, http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Empty(t, w.Body.String())
		mockSvc.AssertExpectations(t)
	})

	t.Run("権限がない場合は 403 を返す", func(t *testing.T) {
		mockSvc := new(mockChannelService)
		mockSvc.On("DeleteChannel", mock.Anything, userID, channelID).Return(apperror.ErrForbidden.WithMessage("You do not have permission"))

		handler := NewChannelHandler(mockSvc)
		router := setupAuthenticatedChannelRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/channels/"+channelID, http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("チャンネルが見つからない場合は 404 を返す", func(t *testing.T) {
		mockSvc := new(mockChannelService)
		mockSvc.On("DeleteChannel", mock.Anything, userID, channelID).Return(apperror.ErrNotFound.WithMessage("Channel not found"))

		handler := NewChannelHandler(mockSvc)
		router := setupAuthenticatedChannelRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/channels/"+channelID, http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("未認証の場合は 401 を返す", func(t *testing.T) {
		mockSvc := new(mockChannelService)
		handler := NewChannelHandler(mockSvc)
		router := setupChannelRouter(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/channels/"+channelID, http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
