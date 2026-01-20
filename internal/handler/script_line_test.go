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

// ScriptLineService のモック
type mockScriptLineService struct {
	mock.Mock
}

func (m *mockScriptLineService) ListByEpisodeID(ctx context.Context, userID, channelID, episodeID string) (*response.ScriptLineListResponse, error) {
	args := m.Called(ctx, userID, channelID, episodeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.ScriptLineListResponse), args.Error(1)
}

func (m *mockScriptLineService) Update(ctx context.Context, userID, channelID, episodeID, lineID string, req request.UpdateScriptLineRequest) (*response.ScriptLineResponse, error) {
	args := m.Called(ctx, userID, channelID, episodeID, lineID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.ScriptLineResponse), args.Error(1)
}

func (m *mockScriptLineService) Delete(ctx context.Context, userID, channelID, episodeID, lineID string) error {
	args := m.Called(ctx, userID, channelID, episodeID, lineID)
	return args.Error(0)
}

func (m *mockScriptLineService) GenerateAudio(ctx context.Context, userID, channelID, episodeID, lineID string) (*response.GenerateAudioResponse, error) {
	args := m.Called(ctx, userID, channelID, episodeID, lineID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.GenerateAudioResponse), args.Error(1)
}

// テスト用のルーターをセットアップする
func setupScriptLineRouter(h *ScriptLineHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/channels/:channelId/episodes/:episodeId/script/lines", h.ListScriptLines)
	return r
}

// 認証済みルーターをセットアップする
func setupAuthenticatedScriptLineRouter(h *ScriptLineHandler, userID string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(string(middleware.UserIDKey), userID)
		c.Next()
	})
	r.GET("/channels/:channelId/episodes/:episodeId/script/lines", h.ListScriptLines)
	return r
}

// テスト用の台本行レスポンスを生成する
func createTestScriptLineResponse() response.ScriptLineResponse {
	now := time.Now()
	emotion := "happy"
	return response.ScriptLineResponse{
		ID:        uuid.New(),
		LineOrder: 1,
		Text:      "テストテキスト",
		Emotion:   &emotion,
		Speaker: response.SpeakerResponse{
			ID:      uuid.New(),
			Name:    "テストスピーカー",
			Persona: "テスト用のペルソナ",
			Voice: response.CharacterVoiceResponse{
				ID:       uuid.New(),
				Name:     "テストボイス",
				Provider: "google",
				Gender:   "female",
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestScriptLineHandler_ListScriptLines(t *testing.T) {
	userID := uuid.New().String()
	channelID := uuid.New().String()
	episodeID := uuid.New().String()

	t.Run("台本行一覧を取得できる", func(t *testing.T) {
		mockSvc := new(mockScriptLineService)
		scriptLines := []response.ScriptLineResponse{createTestScriptLineResponse()}
		result := &response.ScriptLineListResponse{
			Data: scriptLines,
		}
		mockSvc.On("ListByEpisodeID", mock.Anything, userID, channelID, episodeID).Return(result, nil)

		handler := NewScriptLineHandler(mockSvc)
		router := setupAuthenticatedScriptLineRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/channels/"+channelID+"/episodes/"+episodeID+"/script/lines", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp response.ScriptLineListResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Len(t, resp.Data, 1)
		assert.Equal(t, "テストテキスト", resp.Data[0].Text)
		mockSvc.AssertExpectations(t)
	})

	t.Run("空の台本行一覧を取得できる", func(t *testing.T) {
		mockSvc := new(mockScriptLineService)
		result := &response.ScriptLineListResponse{
			Data: []response.ScriptLineResponse{},
		}
		mockSvc.On("ListByEpisodeID", mock.Anything, userID, channelID, episodeID).Return(result, nil)

		handler := NewScriptLineHandler(mockSvc)
		router := setupAuthenticatedScriptLineRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/channels/"+channelID+"/episodes/"+episodeID+"/script/lines", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp response.ScriptLineListResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Empty(t, resp.Data)
		mockSvc.AssertExpectations(t)
	})

	t.Run("チャンネルが見つからない場合は 404 を返す", func(t *testing.T) {
		mockSvc := new(mockScriptLineService)
		mockSvc.On("ListByEpisodeID", mock.Anything, userID, channelID, episodeID).Return(nil, apperror.ErrNotFound.WithMessage("Channel not found"))

		handler := NewScriptLineHandler(mockSvc)
		router := setupAuthenticatedScriptLineRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/channels/"+channelID+"/episodes/"+episodeID+"/script/lines", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("エピソードが見つからない場合は 404 を返す", func(t *testing.T) {
		mockSvc := new(mockScriptLineService)
		mockSvc.On("ListByEpisodeID", mock.Anything, userID, channelID, episodeID).Return(nil, apperror.ErrNotFound.WithMessage("Episode not found"))

		handler := NewScriptLineHandler(mockSvc)
		router := setupAuthenticatedScriptLineRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/channels/"+channelID+"/episodes/"+episodeID+"/script/lines", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("権限がない場合は 403 を返す", func(t *testing.T) {
		mockSvc := new(mockScriptLineService)
		mockSvc.On("ListByEpisodeID", mock.Anything, userID, channelID, episodeID).Return(nil, apperror.ErrForbidden.WithMessage("You do not have permission"))

		handler := NewScriptLineHandler(mockSvc)
		router := setupAuthenticatedScriptLineRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/channels/"+channelID+"/episodes/"+episodeID+"/script/lines", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("サービスがエラーを返すとエラーレスポンスを返す", func(t *testing.T) {
		mockSvc := new(mockScriptLineService)
		mockSvc.On("ListByEpisodeID", mock.Anything, userID, channelID, episodeID).Return(nil, apperror.ErrInternal)

		handler := NewScriptLineHandler(mockSvc)
		router := setupAuthenticatedScriptLineRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/channels/"+channelID+"/episodes/"+episodeID+"/script/lines", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("未認証の場合は 401 を返す", func(t *testing.T) {
		mockSvc := new(mockScriptLineService)
		handler := NewScriptLineHandler(mockSvc)
		router := setupScriptLineRouter(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/channels/"+channelID+"/episodes/"+episodeID+"/script/lines", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
