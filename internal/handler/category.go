package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

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

	Success(c, http.StatusOK, categories)
}

// GetCategoryBySlug godoc
// @Summary カテゴリ取得（スラッグ指定）
// @Description 指定されたスラッグのカテゴリを取得します
// @Tags categories
// @Accept json
// @Produce json
// @Param slug path string true "カテゴリスラッグ"
// @Success 200 {object} response.CategoryDataResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /categories/{slug} [get]
func (h *CategoryHandler) GetCategoryBySlug(c *gin.Context) {
	slug := c.Param("slug")

	category, err := h.categoryService.GetCategoryBySlug(c.Request.Context(), slug)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, http.StatusOK, category)
}
