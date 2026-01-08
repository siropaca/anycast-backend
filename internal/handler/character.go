package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/middleware"
	"github.com/siropaca/anycast-backend/internal/repository"
	"github.com/siropaca/anycast-backend/internal/service"
)

// キャラクター関連のハンドラー
type CharacterHandler struct {
	characterService service.CharacterService
}

// CharacterHandler を作成する
func NewCharacterHandler(cs service.CharacterService) *CharacterHandler {
	return &CharacterHandler{characterService: cs}
}

// ListMyCharacters godoc
// @Summary 自分のキャラクター一覧取得
// @Description 認証ユーザーの所有するキャラクター一覧を取得します
// @Tags me
// @Accept json
// @Produce json
// @Param limit query int false "取得件数（デフォルト: 20、最大: 100）"
// @Param offset query int false "オフセット（デフォルト: 0）"
// @Success 200 {object} response.CharacterListWithPaginationResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /me/characters [get]
func (h *CharacterHandler) ListMyCharacters(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	var req request.ListMyCharactersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(err.Error()))
		return
	}

	filter := repository.CharacterFilter{
		Limit:  req.Limit,
		Offset: req.Offset,
	}

	result, err := h.characterService.ListMyCharacters(c.Request.Context(), userID, filter)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetMyCharacter godoc
// @Summary 自分のキャラクター取得
// @Description 認証ユーザーの所有するキャラクターを取得します
// @Tags me
// @Accept json
// @Produce json
// @Param characterId path string true "キャラクター ID"
// @Success 200 {object} response.CharacterDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /me/characters/{characterId} [get]
func (h *CharacterHandler) GetMyCharacter(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	characterID := c.Param("characterId")
	if characterID == "" {
		Error(c, apperror.ErrValidation.WithMessage("character ID is required"))
		return
	}

	result, err := h.characterService.GetMyCharacter(c.Request.Context(), userID, characterID)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// CreateCharacter godoc
// @Summary キャラクター作成
// @Description 新しいキャラクターを作成します
// @Tags me
// @Accept json
// @Produce json
// @Param request body request.CreateCharacterRequest true "キャラクター作成リクエスト"
// @Success 201 {object} response.CharacterDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse "指定されたボイスまたは画像が見つからない場合"
// @Failure 409 {object} response.ErrorResponse "同じ名前のキャラクターが既に存在する場合"
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /me/characters [post]
func (h *CharacterHandler) CreateCharacter(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	var req request.CreateCharacterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(err.Error()))
		return
	}

	result, err := h.characterService.CreateCharacter(c.Request.Context(), userID, req)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusCreated, result)
}

// UpdateCharacter godoc
// @Summary キャラクター更新
// @Description キャラクターを更新します
// @Tags me
// @Accept json
// @Produce json
// @Param characterId path string true "キャラクター ID"
// @Param request body request.UpdateCharacterRequest true "キャラクター更新リクエスト"
// @Success 200 {object} response.CharacterDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse "キャラクター、ボイス、または画像が見つからない場合"
// @Failure 409 {object} response.ErrorResponse "同じ名前のキャラクターが既に存在する場合"
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /me/characters/{characterId} [patch]
func (h *CharacterHandler) UpdateCharacter(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	characterID := c.Param("characterId")
	if characterID == "" {
		Error(c, apperror.ErrValidation.WithMessage("character ID is required"))
		return
	}

	var req request.UpdateCharacterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(err.Error()))
		return
	}

	result, err := h.characterService.UpdateCharacter(c.Request.Context(), userID, characterID, req)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// DeleteCharacter godoc
// @Summary キャラクター削除
// @Description キャラクターを削除します
// @Tags me
// @Accept json
// @Produce json
// @Param characterId path string true "キャラクター ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse "キャラクターが見つからない場合"
// @Failure 409 {object} response.ErrorResponse "キャラクターが使用中の場合"
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /me/characters/{characterId} [delete]
func (h *CharacterHandler) DeleteCharacter(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	characterID := c.Param("characterId")
	if characterID == "" {
		Error(c, apperror.ErrValidation.WithMessage("character ID is required"))
		return
	}

	if err := h.characterService.DeleteCharacter(c.Request.Context(), userID, characterID); err != nil {
		Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
