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
