package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/middleware"
	"github.com/siropaca/anycast-backend/internal/service"
)

// アートワーク生成関連のハンドラー
type ArtworkHandler struct {
	artworkService service.ArtworkService
}

// ArtworkHandler を作成する
func NewArtworkHandler(as service.ArtworkService) *ArtworkHandler {
	return &ArtworkHandler{artworkService: as}
}

// GenerateChannelArtwork godoc
// @Summary チャンネルアートワーク生成
// @Description チャンネルのアートワークを AI で生成します。プロンプトを省略するとチャンネルのメタデータから自動生成します。
// @Tags channels
// @Accept json
// @Produce json
// @Param channelId path string true "チャンネル ID"
// @Param body body request.GenerateChannelArtworkRequest false "アートワーク生成リクエスト"
// @Success 201 {object} response.ImageUploadDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /channels/{channelId}/artwork/generate [post]
func (h *ArtworkHandler) GenerateChannelArtwork(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	channelID := c.Param("channelId")
	if channelID == "" {
		Error(c, apperror.ErrValidation.WithMessage("channelId は必須です"))
		return
	}

	var req request.GenerateChannelArtworkRequest
	// ボディが空でもエラーにならないよう ShouldBindJSON を使用
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		Error(c, apperror.ErrValidation.WithMessage(formatValidationError(err)))
		return
	}

	result, err := h.artworkService.GenerateChannelArtwork(c.Request.Context(), userID, channelID, req)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusCreated, result)
}

// GenerateEpisodeArtwork godoc
// @Summary エピソードアートワーク生成
// @Description エピソードのアートワークを AI で生成します。プロンプトを省略するとエピソードのメタデータから自動生成します。
// @Tags episodes
// @Accept json
// @Produce json
// @Param channelId path string true "チャンネル ID"
// @Param episodeId path string true "エピソード ID"
// @Param body body request.GenerateEpisodeArtworkRequest false "アートワーク生成リクエスト"
// @Success 201 {object} response.ImageUploadDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /channels/{channelId}/episodes/{episodeId}/artwork/generate [post]
func (h *ArtworkHandler) GenerateEpisodeArtwork(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	channelID := c.Param("channelId")
	if channelID == "" {
		Error(c, apperror.ErrValidation.WithMessage("channelId は必須です"))
		return
	}

	episodeID := c.Param("episodeId")
	if episodeID == "" {
		Error(c, apperror.ErrValidation.WithMessage("episodeId は必須です"))
		return
	}

	var req request.GenerateEpisodeArtworkRequest
	// ボディが空でもエラーにならないよう ShouldBindJSON を使用
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		Error(c, apperror.ErrValidation.WithMessage(formatValidationError(err)))
		return
	}

	result, err := h.artworkService.GenerateEpisodeArtwork(c.Request.Context(), userID, channelID, episodeID, req)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusCreated, result)
}
