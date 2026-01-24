package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/middleware"
	"github.com/siropaca/anycast-backend/internal/service"
)

// BGM 関連のハンドラー
type BgmHandler struct {
	bgmService service.BgmService
}

// BgmHandler を作成する
func NewBgmHandler(bs service.BgmService) *BgmHandler {
	return &BgmHandler{bgmService: bs}
}

// ListMyBgms godoc
// @Summary 自分の BGM 一覧取得
// @Description 認証ユーザーの所有する BGM 一覧を取得します。include_default=true の場合はデフォルト BGM も含めます。
// @Tags me
// @Accept json
// @Produce json
// @Param include_default query bool false "デフォルト BGM を含めるかどうか（デフォルト: false）"
// @Param limit query int false "取得件数（デフォルト: 20、最大: 100）"
// @Param offset query int false "オフセット（デフォルト: 0）"
// @Success 200 {object} response.BgmListWithPaginationResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /me/bgms [get]
func (h *BgmHandler) ListMyBgms(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	var req request.ListMyBgmsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(formatValidationError(err)))
		return
	}

	result, err := h.bgmService.ListMyBgms(c.Request.Context(), userID, req)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// CreateBgm godoc
// @Summary BGM 作成
// @Description 新しい BGM を作成します
// @Tags me
// @Accept json
// @Produce json
// @Param request body request.CreateBgmRequest true "BGM 作成リクエスト"
// @Success 201 {object} response.BgmDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse "指定された音声ファイルが見つからない場合"
// @Failure 409 {object} response.ErrorResponse "同じ名前の BGM が既に存在する場合"
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /me/bgms [post]
func (h *BgmHandler) CreateBgm(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	var req request.CreateBgmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(formatValidationError(err)))
		return
	}

	result, err := h.bgmService.CreateBgm(c.Request.Context(), userID, req)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusCreated, result)
}

// GetMyBgm godoc
// @Summary 自分の BGM 詳細取得
// @Description 認証ユーザーが所有する指定された BGM の詳細を取得します
// @Tags me
// @Accept json
// @Produce json
// @Param bgmId path string true "BGM ID"
// @Success 200 {object} response.BgmDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /me/bgms/{bgmId} [get]
func (h *BgmHandler) GetMyBgm(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	bgmID := c.Param("bgmId")

	result, err := h.bgmService.GetMyBgm(c.Request.Context(), userID, bgmID)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// UpdateMyBgm godoc
// @Summary 自分の BGM 更新
// @Description 認証ユーザーが所有する指定された BGM を更新します
// @Tags me
// @Accept json
// @Produce json
// @Param bgmId path string true "BGM ID"
// @Param request body request.UpdateBgmRequest true "BGM 更新リクエスト"
// @Success 200 {object} response.BgmDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse "同じ名前の BGM が既に存在する場合"
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /me/bgms/{bgmId} [patch]
func (h *BgmHandler) UpdateMyBgm(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	bgmID := c.Param("bgmId")

	var req request.UpdateBgmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(formatValidationError(err)))
		return
	}

	result, err := h.bgmService.UpdateMyBgm(c.Request.Context(), userID, bgmID, req)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// DeleteMyBgm godoc
// @Summary 自分の BGM 削除
// @Description 認証ユーザーが所有する指定された BGM を削除します。エピソードで使用中の場合は削除できません。
// @Tags me
// @Accept json
// @Produce json
// @Param bgmId path string true "BGM ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse "BGM がエピソードで使用中の場合"
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /me/bgms/{bgmId} [delete]
func (h *BgmHandler) DeleteMyBgm(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	bgmID := c.Param("bgmId")

	if err := h.bgmService.DeleteMyBgm(c.Request.Context(), userID, bgmID); err != nil {
		Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
