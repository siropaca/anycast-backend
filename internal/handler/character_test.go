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
	"github.com/siropaca/anycast-backend/internal/repository"
)

// CharacterService のモック
type mockCharacterService struct {
	mock.Mock
}

func (m *mockCharacterService) ListMyCharacters(ctx context.Context, userID string, filter repository.CharacterFilter) (*response.CharacterListWithPaginationResponse, error) {
	args := m.Called(ctx, userID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.CharacterListWithPaginationResponse), args.Error(1)
}

func (m *mockCharacterService) GetMyCharacter(ctx context.Context, userID, characterID string) (*response.CharacterDataResponse, error) {
	args := m.Called(ctx, userID, characterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.CharacterDataResponse), args.Error(1)
}

func (m *mockCharacterService) CreateCharacter(ctx context.Context, userID string, req request.CreateCharacterRequest) (*response.CharacterDataResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.CharacterDataResponse), args.Error(1)
}

func (m *mockCharacterService) UpdateCharacter(ctx context.Context, userID, characterID string, req request.UpdateCharacterRequest) (*response.CharacterDataResponse, error) {
	args := m.Called(ctx, userID, characterID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.CharacterDataResponse), args.Error(1)
}

func (m *mockCharacterService) DeleteCharacter(ctx context.Context, userID, characterID string) error {
	args := m.Called(ctx, userID, characterID)
	return args.Error(0)
}

// ユーザー ID をコンテキストに設定するミドルウェア
func withCharacterUserID(userID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(string(middleware.UserIDKey), userID)
		c.Next()
	}
}

// テスト用のルーターをセットアップする
func setupCharacterRouter(h *CharacterHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/me/characters", h.ListMyCharacters)
	r.POST("/me/characters", h.CreateCharacter)
	r.GET("/me/characters/:characterId", h.GetMyCharacter)
	r.PATCH("/me/characters/:characterId", h.UpdateCharacter)
	r.DELETE("/me/characters/:characterId", h.DeleteCharacter)
	return r
}

// 認証済みルーターをセットアップする
func setupAuthenticatedCharacterRouter(h *CharacterHandler, userID string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(withCharacterUserID(userID))
	r.GET("/me/characters", h.ListMyCharacters)
	r.POST("/me/characters", h.CreateCharacter)
	r.GET("/me/characters/:characterId", h.GetMyCharacter)
	r.PATCH("/me/characters/:characterId", h.UpdateCharacter)
	r.DELETE("/me/characters/:characterId", h.DeleteCharacter)
	return r
}

// テスト用のキャラクターレスポンスを生成する
func createTestCharacterResponse() response.CharacterWithChannelsResponse {
	now := time.Now()
	return response.CharacterWithChannelsResponse{
		ID:      uuid.New(),
		Name:    "Test Character",
		Persona: "A test persona",
		Voice: response.CharacterVoiceResponse{
			ID:       uuid.New(),
			Name:     "Test Voice",
			Provider: "google",
			Gender:   "female",
		},
		Channels:  []response.CharacterChannelResponse{},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestCharacterHandler_ListMyCharacters(t *testing.T) {
	userID := uuid.New().String()

	t.Run("自分のキャラクター一覧を取得できる", func(t *testing.T) {
		mockSvc := new(mockCharacterService)
		charResp := createTestCharacterResponse()
		result := &response.CharacterListWithPaginationResponse{
			Data: []response.CharacterWithChannelsResponse{charResp},
			Pagination: response.PaginationResponse{
				Total:  1,
				Limit:  20,
				Offset: 0,
			},
		}
		mockSvc.On("ListMyCharacters", mock.Anything, userID, mock.AnythingOfType("repository.CharacterFilter")).Return(result, nil)

		handler := NewCharacterHandler(mockSvc)
		router := setupAuthenticatedCharacterRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me/characters", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("サービスがエラーを返すとエラーレスポンスを返す", func(t *testing.T) {
		mockSvc := new(mockCharacterService)
		mockSvc.On("ListMyCharacters", mock.Anything, userID, mock.AnythingOfType("repository.CharacterFilter")).Return(nil, apperror.ErrInternal)

		handler := NewCharacterHandler(mockSvc)
		router := setupAuthenticatedCharacterRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me/characters", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("未認証の場合は 401 を返す", func(t *testing.T) {
		mockSvc := new(mockCharacterService)
		handler := NewCharacterHandler(mockSvc)
		router := setupCharacterRouter(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me/characters", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestCharacterHandler_GetMyCharacter(t *testing.T) {
	userID := uuid.New().String()
	characterID := uuid.New().String()

	t.Run("自分のキャラクターを取得できる", func(t *testing.T) {
		mockSvc := new(mockCharacterService)
		charResp := createTestCharacterResponse()
		result := &response.CharacterDataResponse{Data: charResp}
		mockSvc.On("GetMyCharacter", mock.Anything, userID, characterID).Return(result, nil)

		handler := NewCharacterHandler(mockSvc)
		router := setupAuthenticatedCharacterRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me/characters/"+characterID, http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("キャラクターが見つからない場合は 404 を返す", func(t *testing.T) {
		mockSvc := new(mockCharacterService)
		mockSvc.On("GetMyCharacter", mock.Anything, userID, characterID).Return(nil, apperror.ErrNotFound.WithMessage("キャラクターが見つかりません"))

		handler := NewCharacterHandler(mockSvc)
		router := setupAuthenticatedCharacterRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me/characters/"+characterID, http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("未認証の場合は 401 を返す", func(t *testing.T) {
		mockSvc := new(mockCharacterService)
		handler := NewCharacterHandler(mockSvc)
		router := setupCharacterRouter(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/me/characters/"+characterID, http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestCharacterHandler_CreateCharacter(t *testing.T) {
	userID := uuid.New().String()
	voiceID := uuid.New().String()

	t.Run("キャラクターを作成できる", func(t *testing.T) {
		mockSvc := new(mockCharacterService)
		charResp := createTestCharacterResponse()
		charResp.Name = "New Character"
		result := &response.CharacterDataResponse{Data: charResp}
		mockSvc.On("CreateCharacter", mock.Anything, userID, mock.AnythingOfType("request.CreateCharacterRequest")).Return(result, nil)

		handler := NewCharacterHandler(mockSvc)
		router := setupAuthenticatedCharacterRouter(handler, userID)

		reqBody := request.CreateCharacterRequest{
			Name:    "New Character",
			Persona: "Test persona",
			VoiceID: voiceID,
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/me/characters", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("無効なリクエストの場合は 400 を返す", func(t *testing.T) {
		mockSvc := new(mockCharacterService)
		handler := NewCharacterHandler(mockSvc)
		router := setupAuthenticatedCharacterRouter(handler, userID)

		// 名前なし
		reqBody := map[string]string{
			"persona": "Test persona",
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/me/characters", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockSvc.AssertNotCalled(t, "CreateCharacter")
	})

	t.Run("重複する名前の場合は 409 を返す", func(t *testing.T) {
		mockSvc := new(mockCharacterService)
		mockSvc.On("CreateCharacter", mock.Anything, userID, mock.AnythingOfType("request.CreateCharacterRequest")).Return(nil, apperror.ErrDuplicateName.WithMessage("同じ名前のキャラクターが既に存在します"))

		handler := NewCharacterHandler(mockSvc)
		router := setupAuthenticatedCharacterRouter(handler, userID)

		reqBody := request.CreateCharacterRequest{
			Name:    "Duplicate Name",
			Persona: "Test persona",
			VoiceID: voiceID,
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/me/characters", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("未認証の場合は 401 を返す", func(t *testing.T) {
		mockSvc := new(mockCharacterService)
		handler := NewCharacterHandler(mockSvc)
		router := setupCharacterRouter(handler)

		reqBody := request.CreateCharacterRequest{
			Name:    "New Character",
			Persona: "Test persona",
			VoiceID: voiceID,
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/me/characters", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestCharacterHandler_UpdateCharacter(t *testing.T) {
	userID := uuid.New().String()
	characterID := uuid.New().String()

	t.Run("キャラクターを更新できる", func(t *testing.T) {
		mockSvc := new(mockCharacterService)
		charResp := createTestCharacterResponse()
		charResp.Name = "Updated Character"
		result := &response.CharacterDataResponse{Data: charResp}
		mockSvc.On("UpdateCharacter", mock.Anything, userID, characterID, mock.AnythingOfType("request.UpdateCharacterRequest")).Return(result, nil)

		handler := NewCharacterHandler(mockSvc)
		router := setupAuthenticatedCharacterRouter(handler, userID)

		newName := "Updated Character"
		reqBody := request.UpdateCharacterRequest{
			Name: &newName,
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("PATCH", "/me/characters/"+characterID, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("キャラクターが見つからない場合は 404 を返す", func(t *testing.T) {
		mockSvc := new(mockCharacterService)
		mockSvc.On("UpdateCharacter", mock.Anything, userID, characterID, mock.AnythingOfType("request.UpdateCharacterRequest")).Return(nil, apperror.ErrNotFound.WithMessage("キャラクターが見つかりません"))

		handler := NewCharacterHandler(mockSvc)
		router := setupAuthenticatedCharacterRouter(handler, userID)

		newName := "Updated Character"
		reqBody := request.UpdateCharacterRequest{
			Name: &newName,
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("PATCH", "/me/characters/"+characterID, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("重複する名前の場合は 409 を返す", func(t *testing.T) {
		mockSvc := new(mockCharacterService)
		mockSvc.On("UpdateCharacter", mock.Anything, userID, characterID, mock.AnythingOfType("request.UpdateCharacterRequest")).Return(nil, apperror.ErrDuplicateName.WithMessage("同じ名前のキャラクターが既に存在します"))

		handler := NewCharacterHandler(mockSvc)
		router := setupAuthenticatedCharacterRouter(handler, userID)

		newName := "Duplicate Name"
		reqBody := request.UpdateCharacterRequest{
			Name: &newName,
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("PATCH", "/me/characters/"+characterID, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("未認証の場合は 401 を返す", func(t *testing.T) {
		mockSvc := new(mockCharacterService)
		handler := NewCharacterHandler(mockSvc)
		router := setupCharacterRouter(handler)

		newName := "Updated Character"
		reqBody := request.UpdateCharacterRequest{
			Name: &newName,
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("PATCH", "/me/characters/"+characterID, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestCharacterHandler_DeleteCharacter(t *testing.T) {
	userID := uuid.New().String()
	characterID := uuid.New().String()

	t.Run("キャラクターを削除できる", func(t *testing.T) {
		mockSvc := new(mockCharacterService)
		mockSvc.On("DeleteCharacter", mock.Anything, userID, characterID).Return(nil)

		handler := NewCharacterHandler(mockSvc)
		router := setupAuthenticatedCharacterRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/me/characters/"+characterID, http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Empty(t, w.Body.String())
		mockSvc.AssertExpectations(t)
	})

	t.Run("キャラクターが見つからない場合は 404 を返す", func(t *testing.T) {
		mockSvc := new(mockCharacterService)
		mockSvc.On("DeleteCharacter", mock.Anything, userID, characterID).Return(apperror.ErrNotFound.WithMessage("キャラクターが見つかりません"))

		handler := NewCharacterHandler(mockSvc)
		router := setupAuthenticatedCharacterRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/me/characters/"+characterID, http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("使用中のキャラクターは削除できない", func(t *testing.T) {
		mockSvc := new(mockCharacterService)
		mockSvc.On("DeleteCharacter", mock.Anything, userID, characterID).Return(apperror.ErrCharacterInUse.WithMessage("このキャラクターは使用中のため削除できません"))

		handler := NewCharacterHandler(mockSvc)
		router := setupAuthenticatedCharacterRouter(handler, userID)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/me/characters/"+characterID, http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("未認証の場合は 401 を返す", func(t *testing.T) {
		mockSvc := new(mockCharacterService)
		handler := NewCharacterHandler(mockSvc)
		router := setupCharacterRouter(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/me/characters/"+characterID, http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestNewCharacterHandler(t *testing.T) {
	t.Run("CharacterHandler を作成できる", func(t *testing.T) {
		mockSvc := new(mockCharacterService)
		handler := NewCharacterHandler(mockSvc)
		assert.NotNil(t, handler)
	})
}
