package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// CategoryService のモック
type mockCategoryService struct {
	mock.Mock
}

func (m *mockCategoryService) ListCategories(ctx context.Context) ([]response.CategoryResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]response.CategoryResponse), args.Error(1)
}

func setupCategoryRouter(h *CategoryHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/categories", h.ListCategories)
	return r
}

func TestCategoryHandler_ListCategories(t *testing.T) {
	t.Run("カテゴリ一覧を取得できる", func(t *testing.T) {
		mockSvc := new(mockCategoryService)
		categories := []response.CategoryResponse{
			{ID: uuid.New(), Slug: "technology", Name: "テクノロジー", SortOrder: 0, IsActive: true},
			{ID: uuid.New(), Slug: "news", Name: "ニュース", SortOrder: 1, IsActive: true},
		}
		mockSvc.On("ListCategories", mock.Anything).Return(categories, nil)

		handler := NewCategoryHandler(mockSvc)
		router := setupCategoryRouter(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/categories", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string][]map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Len(t, resp["data"], 2)
		mockSvc.AssertExpectations(t)
	})

	t.Run("空のカテゴリ一覧を取得できる", func(t *testing.T) {
		mockSvc := new(mockCategoryService)
		mockSvc.On("ListCategories", mock.Anything).Return([]response.CategoryResponse{}, nil)

		handler := NewCategoryHandler(mockSvc)
		router := setupCategoryRouter(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/categories", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string][]map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Empty(t, resp["data"])
		mockSvc.AssertExpectations(t)
	})

	t.Run("サービスがエラーを返すとエラーレスポンスを返す", func(t *testing.T) {
		mockSvc := new(mockCategoryService)
		mockSvc.On("ListCategories", mock.Anything).Return(nil, apperror.ErrInternal)

		handler := NewCategoryHandler(mockSvc)
		router := setupCategoryRouter(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/categories", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockSvc.AssertExpectations(t)
	})

	t.Run("画像付きカテゴリを取得できる", func(t *testing.T) {
		mockSvc := new(mockCategoryService)
		imageID := uuid.New()
		categories := []response.CategoryResponse{
			{
				ID:   uuid.New(),
				Slug: "technology",
				Name: "テクノロジー",
				Image: &response.ArtworkResponse{
					ID:  imageID,
					URL: "https://example.com/image.png",
				},
				SortOrder: 0,
				IsActive:  true,
			},
		}
		mockSvc.On("ListCategories", mock.Anything).Return(categories, nil)

		handler := NewCategoryHandler(mockSvc)
		router := setupCategoryRouter(handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/categories", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string][]map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Len(t, resp["data"], 1)
		assert.NotNil(t, resp["data"][0]["image"])
		mockSvc.AssertExpectations(t)
	})
}
