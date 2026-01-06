package handler

import (
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
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/service"
)

// CleanupService のモック
type mockCleanupService struct {
	mock.Mock
}

func (m *mockCleanupService) CleanupOrphanedMedia(ctx context.Context, dryRun bool) (*service.CleanupResult, error) {
	args := m.Called(ctx, dryRun)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.CleanupResult), args.Error(1)
}

// StorageClient のモック
type mockStorageClient struct {
	mock.Mock
}

func (m *mockStorageClient) Upload(ctx context.Context, data []byte, path, contentType string) (string, error) {
	args := m.Called(ctx, data, path, contentType)
	return args.String(0), args.Error(1)
}

func (m *mockStorageClient) GenerateSignedURL(ctx context.Context, path string, expiration time.Duration) (string, error) {
	args := m.Called(ctx, path, expiration)
	return args.String(0), args.Error(1)
}

func (m *mockStorageClient) Delete(ctx context.Context, path string) error {
	args := m.Called(ctx, path)
	return args.Error(0)
}

func setupCleanupRouter(h *CleanupHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/admin/cleanup/orphaned-media", h.CleanupOrphanedMedia)
	return r
}

func TestCleanupHandler_CleanupOrphanedMedia(t *testing.T) {
	t.Run("dry-run モードで孤児メディアを取得できる", func(t *testing.T) {
		mockSvc := new(mockCleanupService)
		mockStorage := new(mockStorageClient)

		audioID := uuid.New()
		imageID := uuid.New()
		now := time.Now()

		result := &service.CleanupResult{
			OrphanedAudios: []model.Audio{
				{ID: audioID, Path: "audios/test.mp3", Filename: "test.mp3", FileSize: 1024, CreatedAt: now},
			},
			OrphanedImages: []model.Image{
				{ID: imageID, URL: "images/test.png", Filename: "test.png", FileSize: 512, CreatedAt: now},
			},
			DeletedAudioCount: 0,
			DeletedImageCount: 0,
			FailedAudioCount:  0,
			FailedImageCount:  0,
		}
		mockSvc.On("CleanupOrphanedMedia", mock.Anything, true).Return(result, nil)
		mockStorage.On("GenerateSignedURL", mock.Anything, "audios/test.mp3", 1*time.Hour).Return("https://signed-url/audio", nil)
		mockStorage.On("GenerateSignedURL", mock.Anything, "images/test.png", 1*time.Hour).Return("https://signed-url/image", nil)

		handler := NewCleanupHandler(mockSvc, mockStorage)
		router := setupCleanupRouter(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/admin/cleanup/orphaned-media?dry_run=true", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		data := resp["data"].(map[string]any)
		assert.True(t, data["dryRun"].(bool))
		assert.Len(t, data["orphanedAudios"], 1)
		assert.Len(t, data["orphanedImages"], 1)
		assert.Equal(t, float64(0), data["deletedAudioCount"])
		assert.Equal(t, float64(0), data["deletedImageCount"])

		mockSvc.AssertExpectations(t)
		mockStorage.AssertExpectations(t)
	})

	t.Run("実行モードで孤児メディアを削除できる", func(t *testing.T) {
		mockSvc := new(mockCleanupService)
		mockStorage := new(mockStorageClient)

		result := &service.CleanupResult{
			OrphanedAudios:    []model.Audio{},
			OrphanedImages:    []model.Image{},
			DeletedAudioCount: 2,
			DeletedImageCount: 1,
			FailedAudioCount:  0,
			FailedImageCount:  0,
		}
		mockSvc.On("CleanupOrphanedMedia", mock.Anything, false).Return(result, nil)

		handler := NewCleanupHandler(mockSvc, mockStorage)
		router := setupCleanupRouter(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/admin/cleanup/orphaned-media", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		data := resp["data"].(map[string]any)
		assert.False(t, data["dryRun"].(bool))
		assert.Equal(t, float64(2), data["deletedAudioCount"])
		assert.Equal(t, float64(1), data["deletedImageCount"])

		mockSvc.AssertExpectations(t)
	})

	t.Run("dry_run=false を明示的に指定できる", func(t *testing.T) {
		mockSvc := new(mockCleanupService)
		mockStorage := new(mockStorageClient)

		result := &service.CleanupResult{
			OrphanedAudios:    []model.Audio{},
			OrphanedImages:    []model.Image{},
			DeletedAudioCount: 0,
			DeletedImageCount: 0,
			FailedAudioCount:  0,
			FailedImageCount:  0,
		}
		mockSvc.On("CleanupOrphanedMedia", mock.Anything, false).Return(result, nil)

		handler := NewCleanupHandler(mockSvc, mockStorage)
		router := setupCleanupRouter(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/admin/cleanup/orphaned-media?dry_run=false", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("サービスがエラーを返すとエラーレスポンスを返す", func(t *testing.T) {
		mockSvc := new(mockCleanupService)
		mockStorage := new(mockStorageClient)

		mockSvc.On("CleanupOrphanedMedia", mock.Anything, false).Return(nil, apperror.ErrInternal)

		handler := NewCleanupHandler(mockSvc, mockStorage)
		router := setupCleanupRouter(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/admin/cleanup/orphaned-media", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("storage client が nil でも動作する", func(t *testing.T) {
		mockSvc := new(mockCleanupService)

		audioID := uuid.New()
		now := time.Now()

		result := &service.CleanupResult{
			OrphanedAudios: []model.Audio{
				{ID: audioID, Path: "audios/test.mp3", Filename: "test.mp3", FileSize: 1024, CreatedAt: now},
			},
			OrphanedImages:    []model.Image{},
			DeletedAudioCount: 0,
			DeletedImageCount: 0,
			FailedAudioCount:  0,
			FailedImageCount:  0,
		}
		mockSvc.On("CleanupOrphanedMedia", mock.Anything, true).Return(result, nil)

		handler := NewCleanupHandler(mockSvc, nil)
		router := setupCleanupRouter(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/admin/cleanup/orphaned-media?dry_run=true", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		data := resp["data"].(map[string]any)
		audios := data["orphanedAudios"].([]any)
		assert.Len(t, audios, 1)
		// URL は空文字になる
		assert.Equal(t, "", audios[0].(map[string]any)["url"])

		mockSvc.AssertExpectations(t)
	})
}

func TestToOrphanedAudioResponses(t *testing.T) {
	t.Run("空のスライスを変換すると空のスライスを返す", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", http.NoBody)

		audios := []model.Audio{}
		resp := toOrphanedAudioResponses(audios, nil, c)

		assert.Empty(t, resp)
	})

	t.Run("複数の Audio を変換できる", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", http.NoBody)

		id1 := uuid.New()
		id2 := uuid.New()
		now := time.Now()

		audios := []model.Audio{
			{ID: id1, Path: "audios/1.mp3", Filename: "1.mp3", FileSize: 100, CreatedAt: now},
			{ID: id2, Path: "audios/2.mp3", Filename: "2.mp3", FileSize: 200, CreatedAt: now},
		}

		resp := toOrphanedAudioResponses(audios, nil, c)

		assert.Len(t, resp, 2)
		assert.Equal(t, id1, resp[0].ID)
		assert.Equal(t, "1.mp3", resp[0].Filename)
		assert.Equal(t, id2, resp[1].ID)
		assert.Equal(t, "2.mp3", resp[1].Filename)
	})
}

func TestToOrphanedImageResponses(t *testing.T) {
	t.Run("空のスライスを変換すると空のスライスを返す", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", http.NoBody)

		images := []model.Image{}
		resp := toOrphanedImageResponses(images, nil, c)

		assert.Empty(t, resp)
	})

	t.Run("複数の Image を変換できる", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", http.NoBody)

		id1 := uuid.New()
		id2 := uuid.New()
		now := time.Now()

		images := []model.Image{
			{ID: id1, URL: "images/1.png", Filename: "1.png", FileSize: 100, CreatedAt: now},
			{ID: id2, URL: "images/2.png", Filename: "2.png", FileSize: 200, CreatedAt: now},
		}

		resp := toOrphanedImageResponses(images, nil, c)

		assert.Len(t, resp, 2)
		assert.Equal(t, id1, resp[0].ID)
		assert.Equal(t, "1.png", resp[0].Filename)
		assert.Equal(t, id2, resp[1].ID)
		assert.Equal(t, "2.png", resp[1].Filename)
	})
}

func TestToOrphanedAudioResponse(t *testing.T) {
	t.Run("署名付き URL を生成できる", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", http.NoBody)

		mockStorage := new(mockStorageClient)
		mockStorage.On("GenerateSignedURL", mock.Anything, "audios/test.mp3", 1*time.Hour).Return("https://signed-url", nil)

		id := uuid.New()
		now := time.Now()
		audio := &model.Audio{ID: id, Path: "audios/test.mp3", Filename: "test.mp3", FileSize: 1024, CreatedAt: now}

		resp := toOrphanedAudioResponse(audio, mockStorage, c)

		assert.Equal(t, id, resp.ID)
		assert.Equal(t, "https://signed-url", resp.URL)
		assert.Equal(t, "test.mp3", resp.Filename)
		assert.Equal(t, 1024, resp.FileSize)
		assert.Equal(t, now, resp.CreatedAt)

		mockStorage.AssertExpectations(t)
	})

	t.Run("署名付き URL の生成が失敗しても空文字を返す", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", http.NoBody)

		mockStorage := new(mockStorageClient)
		mockStorage.On("GenerateSignedURL", mock.Anything, "audios/test.mp3", 1*time.Hour).Return("", apperror.ErrInternal)

		audio := &model.Audio{ID: uuid.New(), Path: "audios/test.mp3", Filename: "test.mp3", FileSize: 1024}

		resp := toOrphanedAudioResponse(audio, mockStorage, c)

		assert.Equal(t, "", resp.URL)
		mockStorage.AssertExpectations(t)
	})
}

func TestToOrphanedImageResponse(t *testing.T) {
	t.Run("署名付き URL を生成できる", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", http.NoBody)

		mockStorage := new(mockStorageClient)
		mockStorage.On("GenerateSignedURL", mock.Anything, "images/test.png", 1*time.Hour).Return("https://signed-url", nil)

		id := uuid.New()
		now := time.Now()
		image := &model.Image{ID: id, URL: "images/test.png", Filename: "test.png", FileSize: 512, CreatedAt: now}

		resp := toOrphanedImageResponse(image, mockStorage, c)

		assert.Equal(t, id, resp.ID)
		assert.Equal(t, "https://signed-url", resp.URL)
		assert.Equal(t, "test.png", resp.Filename)
		assert.Equal(t, 512, resp.FileSize)
		assert.Equal(t, now, resp.CreatedAt)

		mockStorage.AssertExpectations(t)
	})

	t.Run("storage client が nil の場合は空文字を返す", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", http.NoBody)

		image := &model.Image{ID: uuid.New(), URL: "images/test.png", Filename: "test.png", FileSize: 512}

		resp := toOrphanedImageResponse(image, nil, c)

		assert.Equal(t, "", resp.URL)
	})
}
