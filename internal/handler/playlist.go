package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/middleware"
	"github.com/siropaca/anycast-backend/internal/service"
)

// プレイリスト関連のハンドラー
type PlaylistHandler struct {
	playlistService service.PlaylistService
}

// PlaylistHandler を作成する
func NewPlaylistHandler(ps service.PlaylistService) *PlaylistHandler {
	return &PlaylistHandler{playlistService: ps}
}

// ListPlaylists godoc
// @Summary 自分のプレイリスト一覧取得
// @Description 自分のプレイリスト一覧を取得します
// @Tags me
// @Accept json
// @Produce json
// @Param limit query int false "取得件数（デフォルト: 20、最大: 100）"
// @Param offset query int false "オフセット（デフォルト: 0）"
// @Success 200 {object} response.PlaylistListWithPaginationResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /me/playlists [get]
func (h *PlaylistHandler) ListPlaylists(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	var req request.ListPlaylistsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(formatValidationError(err)))
		return
	}

	result, err := h.playlistService.ListPlaylists(c.Request.Context(), userID, req.Limit, req.Offset)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetPlaylist godoc
// @Summary 自分のプレイリスト詳細取得
// @Description 自分のプレイリスト詳細を取得します（アイテム含む）
// @Tags me
// @Accept json
// @Produce json
// @Param playlistId path string true "プレイリスト ID"
// @Success 200 {object} response.PlaylistDetailDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /me/playlists/{playlistId} [get]
func (h *PlaylistHandler) GetPlaylist(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	playlistID := c.Param("playlistId")
	if playlistID == "" {
		Error(c, apperror.ErrValidation.WithMessage("playlistId は必須です"))
		return
	}

	result, err := h.playlistService.GetPlaylist(c.Request.Context(), userID, playlistID)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// CreatePlaylist godoc
// @Summary プレイリスト作成
// @Description 新しいプレイリストを作成します
// @Tags me
// @Accept json
// @Produce json
// @Param request body request.CreatePlaylistRequest true "プレイリスト作成リクエスト"
// @Success 201 {object} response.PlaylistDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /me/playlists [post]
func (h *PlaylistHandler) CreatePlaylist(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	var req request.CreatePlaylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(formatValidationError(err)))
		return
	}

	result, err := h.playlistService.CreatePlaylist(c.Request.Context(), userID, req)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusCreated, result)
}

// UpdatePlaylist godoc
// @Summary プレイリスト更新
// @Description 指定したプレイリストを更新します
// @Tags me
// @Accept json
// @Produce json
// @Param playlistId path string true "プレイリスト ID"
// @Param request body request.UpdatePlaylistRequest true "プレイリスト更新リクエスト"
// @Success 200 {object} response.PlaylistDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /me/playlists/{playlistId} [patch]
func (h *PlaylistHandler) UpdatePlaylist(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	playlistID := c.Param("playlistId")
	if playlistID == "" {
		Error(c, apperror.ErrValidation.WithMessage("playlistId は必須です"))
		return
	}

	var req request.UpdatePlaylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(formatValidationError(err)))
		return
	}

	result, err := h.playlistService.UpdatePlaylist(c.Request.Context(), userID, playlistID, req)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// DeletePlaylist godoc
// @Summary プレイリスト削除
// @Description 指定したプレイリストを削除します（デフォルトプレイリストは削除不可）
// @Tags me
// @Accept json
// @Produce json
// @Param playlistId path string true "プレイリスト ID"
// @Success 204 "No Content"
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /me/playlists/{playlistId} [delete]
func (h *PlaylistHandler) DeletePlaylist(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	playlistID := c.Param("playlistId")
	if playlistID == "" {
		Error(c, apperror.ErrValidation.WithMessage("playlistId は必須です"))
		return
	}

	if err := h.playlistService.DeletePlaylist(c.Request.Context(), userID, playlistID); err != nil {
		Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// ReorderItems godoc
// @Summary プレイリストアイテム並び替え
// @Description プレイリスト内のアイテムの順序を変更します
// @Tags me
// @Accept json
// @Produce json
// @Param playlistId path string true "プレイリスト ID"
// @Param request body request.ReorderPlaylistItemsRequest true "並び替えリクエスト"
// @Success 200 {object} response.PlaylistDetailDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /me/playlists/{playlistId}/items/reorder [post]
func (h *PlaylistHandler) ReorderItems(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	playlistID := c.Param("playlistId")
	if playlistID == "" {
		Error(c, apperror.ErrValidation.WithMessage("playlistId は必須です"))
		return
	}

	var req request.ReorderPlaylistItemsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(formatValidationError(err)))
		return
	}

	result, err := h.playlistService.ReorderItems(c.Request.Context(), userID, playlistID, req)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// UpdateEpisodePlaylists godoc
// @Summary エピソードのプレイリスト所属一括更新
// @Description エピソードが所属するプレイリストを一括更新します
// @Tags episodes
// @Accept json
// @Produce json
// @Param episodeId path string true "エピソード ID"
// @Param request body request.UpdateEpisodePlaylistsRequest true "プレイリスト所属更新リクエスト"
// @Success 200 {object} response.EpisodePlaylistIDsDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /episodes/{episodeId}/playlists [put]
func (h *PlaylistHandler) UpdateEpisodePlaylists(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	episodeID := c.Param("episodeId")
	if episodeID == "" {
		Error(c, apperror.ErrValidation.WithMessage("episodeId は必須です"))
		return
	}

	var req request.UpdateEpisodePlaylistsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(formatValidationError(err)))
		return
	}

	result, err := h.playlistService.UpdateEpisodePlaylists(c.Request.Context(), userID, episodeID, req)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}
