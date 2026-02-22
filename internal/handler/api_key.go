package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/middleware"
	"github.com/siropaca/anycast-backend/internal/service"
)

// API キー関連のハンドラー
type APIKeyHandler struct {
	apiKeyService service.APIKeyService
}

// APIKeyHandler を作成する
func NewAPIKeyHandler(aks service.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{
		apiKeyService: aks,
	}
}

// CreateAPIKey godoc
// @Summary API キー作成
// @Description 新しい API キーを発行します。平文キーはこのレスポンスでのみ返却されます。
// @Tags api-keys
// @Accept json
// @Produce json
// @Param request body request.CreateAPIKeyRequest true "API キー作成情報"
// @Success 201 {object} response.APIKeyCreatedDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /me/api-keys [post]
func (h *APIKeyHandler) CreateAPIKey(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	var req request.CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(formatValidationError(err)))
		return
	}

	resp, err := h.apiKeyService.Create(c.Request.Context(), userID, req)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, http.StatusCreated, resp)
}

// ListAPIKeys godoc
// @Summary API キー一覧取得
// @Description 自分の API キー一覧を取得します。平文キーは含まれません。
// @Tags api-keys
// @Produce json
// @Success 200 {object} response.APIKeyListDataResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /me/api-keys [get]
func (h *APIKeyHandler) ListAPIKeys(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	apiKeys, err := h.apiKeyService.List(c.Request.Context(), userID)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": apiKeys})
}

// DeleteAPIKey godoc
// @Summary API キー削除
// @Description 指定した API キーを削除します。
// @Tags api-keys
// @Param apiKeyId path string true "API キー ID"
// @Success 204
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /me/api-keys/{apiKeyId} [delete]
func (h *APIKeyHandler) DeleteAPIKey(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	apiKeyID := c.Param("apiKeyId")

	if err := h.apiKeyService.Delete(c.Request.Context(), userID, apiKeyID); err != nil {
		Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// ListAPIKeys で使用するレスポンス型のコンパイル時チェック
var _ = response.APIKeyListDataResponse{}
