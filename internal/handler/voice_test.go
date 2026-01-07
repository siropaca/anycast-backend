package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// VoiceService のモック
type mockVoiceService struct {
	mock.Mock
}

func (m *mockVoiceService) ListVoices(ctx context.Context, filter repository.VoiceFilter) ([]model.Voice, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Voice), args.Error(1)
}

func (m *mockVoiceService) GetVoice(ctx context.Context, id string) (*model.Voice, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Voice), args.Error(1)
}

func setupRouter(h *VoiceHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/voices", h.ListVoices)
	r.GET("/voices/:voiceId", h.GetVoice)
	return r
}

func TestVoiceHandler_ListVoices(t *testing.T) {
	t.Run("ボイス一覧を取得できる", func(t *testing.T) {
		mockSvc := new(mockVoiceService)
		voices := []model.Voice{
			{ID: uuid.New(), Provider: "google", Name: "Voice 1", Gender: "female", IsActive: true},
		}
		mockSvc.On("ListVoices", mock.Anything, repository.VoiceFilter{}).Return(voices, nil)

		handler := NewVoiceHandler(mockSvc)
		router := setupRouter(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/voices", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string][]map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Len(t, resp["data"], 1)
		mockSvc.AssertExpectations(t)
	})

	t.Run("クエリパラメータでフィルタできる", func(t *testing.T) {
		mockSvc := new(mockVoiceService)
		provider := "google"
		gender := "female"
		filter := repository.VoiceFilter{Provider: &provider, Gender: &gender}
		voices := []model.Voice{
			{ID: uuid.New(), Provider: "google", Gender: "female"},
		}
		mockSvc.On("ListVoices", mock.Anything, filter).Return(voices, nil)

		handler := NewVoiceHandler(mockSvc)
		router := setupRouter(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/voices?provider=google&gender=female", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("サービスがエラーを返すとエラーレスポンスを返す", func(t *testing.T) {
		mockSvc := new(mockVoiceService)
		mockSvc.On("ListVoices", mock.Anything, repository.VoiceFilter{}).Return(nil, apperror.ErrInternal)

		handler := NewVoiceHandler(mockSvc)
		router := setupRouter(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/voices", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})
}

func TestVoiceHandler_GetVoice(t *testing.T) {
	t.Run("ID でボイスを取得できる", func(t *testing.T) {
		mockSvc := new(mockVoiceService)
		id := uuid.New()
		voice := &model.Voice{ID: id, Provider: "google", Name: "Test Voice", Gender: "male", IsActive: true}
		mockSvc.On("GetVoice", mock.Anything, id.String()).Return(voice, nil)

		handler := NewVoiceHandler(mockSvc)
		router := setupRouter(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/voices/"+id.String(), http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, id.String(), resp["data"]["id"])
		mockSvc.AssertExpectations(t)
	})

	t.Run("無効な UUID 形式の場合は 400 を返す", func(t *testing.T) {
		mockSvc := new(mockVoiceService)

		handler := NewVoiceHandler(mockSvc)
		router := setupRouter(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/voices/invalid-uuid", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockSvc.AssertNotCalled(t, "GetVoice")
	})

	t.Run("存在しないボイスの場合は 404 を返す", func(t *testing.T) {
		mockSvc := new(mockVoiceService)
		id := uuid.New()
		mockSvc.On("GetVoice", mock.Anything, id.String()).Return(nil, apperror.ErrNotFound)

		handler := NewVoiceHandler(mockSvc)
		router := setupRouter(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/voices/"+id.String(), http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("サービスがエラーを返すとエラーレスポンスを返す", func(t *testing.T) {
		mockSvc := new(mockVoiceService)
		id := uuid.New()
		mockSvc.On("GetVoice", mock.Anything, id.String()).Return(nil, errors.New("unexpected error"))

		handler := NewVoiceHandler(mockSvc)
		router := setupRouter(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/voices/"+id.String(), http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})
}

func TestToVoiceResponse(t *testing.T) {
	t.Run("Voice モデルを VoiceResponse に変換できる", func(t *testing.T) {
		id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
		voice := &model.Voice{
			ID:              id,
			Provider:        "google",
			ProviderVoiceID: "en-US-Neural2-A",
			Name:            "American English Female",
			Gender:          "female",
			IsActive:        true,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		resp := toVoiceResponse(voice)

		assert.Equal(t, id, resp.ID)
		assert.Equal(t, "google", resp.Provider)
		assert.Equal(t, "en-US-Neural2-A", resp.ProviderVoiceID)
		assert.Equal(t, "American English Female", resp.Name)
		assert.Equal(t, "female", resp.Gender)
		assert.True(t, resp.IsActive)
	})

	t.Run("IsActive が false の場合も正しく変換される", func(t *testing.T) {
		voice := &model.Voice{
			ID:       uuid.New(),
			IsActive: false,
		}

		resp := toVoiceResponse(voice)

		assert.False(t, resp.IsActive)
	})
}

func TestToVoiceResponses(t *testing.T) {
	t.Run("空のスライスを変換すると空のスライスを返す", func(t *testing.T) {
		voices := []model.Voice{}

		resp := toVoiceResponses(voices)

		assert.Empty(t, resp)
	})

	t.Run("複数の Voice を変換できる", func(t *testing.T) {
		id1 := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
		id2 := uuid.MustParse("550e8400-e29b-41d4-a716-446655440002")
		voices := []model.Voice{
			{
				ID:              id1,
				Provider:        "google",
				ProviderVoiceID: "voice-1",
				Name:            "Voice 1",
				Gender:          "male",
				IsActive:        true,
			},
			{
				ID:              id2,
				Provider:        "amazon",
				ProviderVoiceID: "voice-2",
				Name:            "Voice 2",
				Gender:          "female",
				IsActive:        false,
			},
		}

		resp := toVoiceResponses(voices)

		assert.Len(t, resp, 2)
		assert.Equal(t, id1, resp[0].ID)
		assert.Equal(t, "google", resp[0].Provider)
		assert.Equal(t, id2, resp[1].ID)
		assert.Equal(t, "amazon", resp[1].Provider)
	})

	t.Run("変換結果の長さが入力と一致する", func(t *testing.T) {
		voices := make([]model.Voice, 5)
		for i := range voices {
			voices[i] = model.Voice{ID: uuid.New()}
		}

		resp := toVoiceResponses(voices)

		assert.Len(t, resp, len(voices))
	})
}
