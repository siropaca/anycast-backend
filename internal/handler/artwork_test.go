package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/middleware"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// ArtworkService のモック
type mockArtworkService struct {
	mock.Mock
}

func (m *mockArtworkService) GenerateChannelArtwork(ctx context.Context, userID, channelID string, req request.GenerateChannelArtworkRequest) (*response.ImageUploadDataResponse, error) {
	args := m.Called(ctx, userID, channelID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.ImageUploadDataResponse), args.Error(1)
}

func (m *mockArtworkService) GenerateEpisodeArtwork(ctx context.Context, userID, channelID, episodeID string, req request.GenerateEpisodeArtworkRequest) (*response.ImageUploadDataResponse, error) {
	args := m.Called(ctx, userID, channelID, episodeID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.ImageUploadDataResponse), args.Error(1)
}

func setupArtworkRouter(service *mockArtworkService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewArtworkHandler(service)

	// 認証済みユーザーをシミュレートするミドルウェア
	authMiddleware := func(userID string) gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set(string(middleware.UserIDKey), userID)
			c.Next()
		}
	}

	r.POST("/channels/:channelId/artwork/generate", authMiddleware("user-123"), handler.GenerateChannelArtwork)
	r.POST("/channels/:channelId/episodes/:episodeId/artwork/generate", authMiddleware("user-123"), handler.GenerateEpisodeArtwork)

	// 認証なしルート
	r.POST("/no-auth/channels/:channelId/artwork/generate", handler.GenerateChannelArtwork)

	return r
}

func TestArtworkHandler_GenerateChannelArtwork(t *testing.T) {
	channelID := uuid.New()
	imageID := uuid.New()

	t.Run("プロンプト指定でチャンネルアートワークを生成できる", func(t *testing.T) {
		mockSvc := new(mockArtworkService)
		result := &response.ImageUploadDataResponse{
			Data: response.ImageUploadResponse{
				ID:       imageID,
				MimeType: "image/png",
				URL:      "https://example.com/artwork.png",
				Filename: "artwork_12345678.png",
				FileSize: 1234567,
			},
		}
		mockSvc.On("GenerateChannelArtwork", mock.Anything, "user-123", channelID.String(), mock.Anything).Return(result, nil)

		router := setupArtworkRouter(mockSvc)

		body := `{"prompt":"テスト用プロンプト","setArtwork":true}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/channels/"+channelID.String()+"/artwork/generate", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp response.ImageUploadDataResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, imageID, resp.Data.ID)
		assert.Equal(t, "image/png", resp.Data.MimeType)
		mockSvc.AssertExpectations(t)
	})

	t.Run("空ボディでチャンネルアートワークを生成できる", func(t *testing.T) {
		mockSvc := new(mockArtworkService)
		result := &response.ImageUploadDataResponse{
			Data: response.ImageUploadResponse{
				ID:       imageID,
				MimeType: "image/png",
				URL:      "https://example.com/artwork.png",
				Filename: "artwork_12345678.png",
				FileSize: 1234567,
			},
		}
		mockSvc.On("GenerateChannelArtwork", mock.Anything, "user-123", channelID.String(), mock.Anything).Return(result, nil)

		router := setupArtworkRouter(mockSvc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/channels/"+channelID.String()+"/artwork/generate", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("未認証の場合は 401 を返す", func(t *testing.T) {
		mockSvc := new(mockArtworkService)
		router := setupArtworkRouter(mockSvc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/no-auth/channels/"+channelID.String()+"/artwork/generate", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockSvc.AssertNotCalled(t, "GenerateChannelArtwork")
	})

	t.Run("バリデーションエラーの場合は 400 を返す", func(t *testing.T) {
		mockSvc := new(mockArtworkService)
		router := setupArtworkRouter(mockSvc)

		// 1000 文字を超えるプロンプト
		longPrompt := strings.Repeat("あ", 1001)
		body := `{"prompt":"` + longPrompt + `"}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/channels/"+channelID.String()+"/artwork/generate", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockSvc.AssertNotCalled(t, "GenerateChannelArtwork")
	})

	t.Run("サービスが 403 を返す場合は 403 を返す", func(t *testing.T) {
		mockSvc := new(mockArtworkService)
		mockSvc.On("GenerateChannelArtwork", mock.Anything, "user-123", channelID.String(), mock.Anything).
			Return(nil, apperror.ErrForbidden.WithMessage("このチャンネルのアートワーク生成権限がありません"))

		router := setupArtworkRouter(mockSvc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/channels/"+channelID.String()+"/artwork/generate", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("サービスが内部エラーを返すと 500 を返す", func(t *testing.T) {
		mockSvc := new(mockArtworkService)
		mockSvc.On("GenerateChannelArtwork", mock.Anything, "user-123", channelID.String(), mock.Anything).
			Return(nil, apperror.ErrGenerationFailed.WithMessage("画像生成に失敗しました"))

		router := setupArtworkRouter(mockSvc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/channels/"+channelID.String()+"/artwork/generate", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})
}

func TestArtworkHandler_GenerateEpisodeArtwork(t *testing.T) {
	channelID := uuid.New()
	episodeID := uuid.New()
	imageID := uuid.New()

	t.Run("プロンプト指定でエピソードアートワークを生成できる", func(t *testing.T) {
		mockSvc := new(mockArtworkService)
		result := &response.ImageUploadDataResponse{
			Data: response.ImageUploadResponse{
				ID:       imageID,
				MimeType: "image/png",
				URL:      "https://example.com/artwork.png",
				Filename: "artwork_12345678.png",
				FileSize: 1234567,
			},
		}
		mockSvc.On("GenerateEpisodeArtwork", mock.Anything, "user-123", channelID.String(), episodeID.String(), mock.Anything).Return(result, nil)

		router := setupArtworkRouter(mockSvc)

		body := `{"prompt":"エピソードアートワーク","setArtwork":false}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/channels/"+channelID.String()+"/episodes/"+episodeID.String()+"/artwork/generate", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp response.ImageUploadDataResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, imageID, resp.Data.ID)
		mockSvc.AssertExpectations(t)
	})

	t.Run("空ボディでエピソードアートワークを生成できる", func(t *testing.T) {
		mockSvc := new(mockArtworkService)
		result := &response.ImageUploadDataResponse{
			Data: response.ImageUploadResponse{
				ID:       imageID,
				MimeType: "image/png",
				URL:      "https://example.com/artwork.png",
				Filename: "artwork_12345678.png",
				FileSize: 1234567,
			},
		}
		mockSvc.On("GenerateEpisodeArtwork", mock.Anything, "user-123", channelID.String(), episodeID.String(), mock.Anything).Return(result, nil)

		router := setupArtworkRouter(mockSvc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/channels/"+channelID.String()+"/episodes/"+episodeID.String()+"/artwork/generate", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("サービスが 404 を返す場合は 404 を返す", func(t *testing.T) {
		mockSvc := new(mockArtworkService)
		mockSvc.On("GenerateEpisodeArtwork", mock.Anything, "user-123", channelID.String(), episodeID.String(), mock.Anything).
			Return(nil, apperror.ErrNotFound.WithMessage("エピソードが見つかりません"))

		router := setupArtworkRouter(mockSvc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/channels/"+channelID.String()+"/episodes/"+episodeID.String()+"/artwork/generate", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockSvc.AssertExpectations(t)
	})
}

func TestNewArtworkHandler(t *testing.T) {
	t.Run("ArtworkHandler を作成できる", func(t *testing.T) {
		mockSvc := new(mockArtworkService)
		handler := NewArtworkHandler(mockSvc)
		assert.NotNil(t, handler)
	})
}
