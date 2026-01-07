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
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// CategoryService のモック
type mockCategoryService struct {
	mock.Mock
}

func (m *mockCategoryService) ListCategories(ctx context.Context) ([]model.Category, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Category), args.Error(1)
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
		categories := []model.Category{
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
		mockSvc.On("ListCategories", mock.Anything).Return([]model.Category{}, nil)

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
}

func TestToCategoryResponse(t *testing.T) {
	t.Run("Category モデルを CategoryResponse に変換できる", func(t *testing.T) {
		id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
		category := &model.Category{
			ID:        id,
			Slug:      "technology",
			Name:      "テクノロジー",
			SortOrder: 5,
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		resp := toCategoryResponse(category)

		assert.Equal(t, id, resp.ID)
		assert.Equal(t, "technology", resp.Slug)
		assert.Equal(t, "テクノロジー", resp.Name)
		assert.Equal(t, 5, resp.SortOrder)
		assert.True(t, resp.IsActive)
	})

	t.Run("SortOrder が 0 の場合も正しく変換される", func(t *testing.T) {
		category := &model.Category{
			ID:        uuid.New(),
			Slug:      "first",
			Name:      "First",
			SortOrder: 0,
			IsActive:  true,
		}

		resp := toCategoryResponse(category)

		assert.Equal(t, 0, resp.SortOrder)
	})

	t.Run("IsActive が false の場合も正しく変換される", func(t *testing.T) {
		category := &model.Category{
			ID:        uuid.New(),
			Slug:      "inactive",
			Name:      "Inactive",
			SortOrder: 0,
			IsActive:  false,
		}

		resp := toCategoryResponse(category)

		assert.False(t, resp.IsActive)
	})
}

func TestToCategoryResponses(t *testing.T) {
	t.Run("空のスライスを変換すると空のスライスを返す", func(t *testing.T) {
		categories := []model.Category{}

		resp := toCategoryResponses(categories)

		assert.Empty(t, resp)
	})

	t.Run("複数の Category を変換できる", func(t *testing.T) {
		id1 := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
		id2 := uuid.MustParse("550e8400-e29b-41d4-a716-446655440002")
		categories := []model.Category{
			{
				ID:        id1,
				Slug:      "technology",
				Name:      "テクノロジー",
				SortOrder: 0,
				IsActive:  true,
			},
			{
				ID:        id2,
				Slug:      "news",
				Name:      "ニュース",
				SortOrder: 1,
				IsActive:  false,
			},
		}

		resp := toCategoryResponses(categories)

		assert.Len(t, resp, 2)
		assert.Equal(t, id1, resp[0].ID)
		assert.Equal(t, "technology", resp[0].Slug)
		assert.Equal(t, id2, resp[1].ID)
		assert.Equal(t, "news", resp[1].Slug)
	})

	t.Run("変換結果の長さが入力と一致する", func(t *testing.T) {
		categories := make([]model.Category, 5)
		for i := range categories {
			categories[i] = model.Category{ID: uuid.New(), Slug: "slug-" + string(rune(i)), SortOrder: i, IsActive: true}
		}

		resp := toCategoryResponses(categories)

		assert.Len(t, resp, len(categories))
	})
}
