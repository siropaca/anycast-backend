package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/service"
)

// カテゴリ関連のハンドラー
type CategoryHandler struct {
	categoryService service.CategoryService
}

// CategoryHandler を作成する
func NewCategoryHandler(cs service.CategoryService) *CategoryHandler {
	return &CategoryHandler{categoryService: cs}
}

// ListCategories godoc
// @Summary カテゴリ一覧取得
// @Description カテゴリの一覧を取得します
// @Tags categories
// @Accept json
// @Produce json
// @Success 200 {object} response.CategoryListResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /categories [get]
func (h *CategoryHandler) ListCategories(c *gin.Context) {
	categories, err := h.categoryService.ListCategories(c.Request.Context())
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, http.StatusOK, toCategoryResponses(categories))
}

// Category モデルのスライスをレスポンス DTO のスライスに変換する
func toCategoryResponses(categories []model.Category) []response.CategoryResponse {
	result := make([]response.CategoryResponse, len(categories))

	for i, c := range categories {
		result[i] = toCategoryResponse(&c)
	}

	return result
}

// Category モデルをレスポンス DTO に変換する
func toCategoryResponse(c *model.Category) response.CategoryResponse {
	return response.CategoryResponse{
		ID:        c.ID,
		Slug:      c.Slug,
		Name:      c.Name,
		SortOrder: c.SortOrder,
		IsActive:  c.IsActive,
	}
}
