package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/middleware"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
	"github.com/siropaca/anycast-backend/internal/service"
)

// ボイス関連のハンドラー
type VoiceHandler struct {
	voiceService service.VoiceService
}

// VoiceHandler を作成する
func NewVoiceHandler(vs service.VoiceService) *VoiceHandler {
	return &VoiceHandler{voiceService: vs}
}

// ListVoices godoc
// @Summary ボイス一覧取得
// @Description 利用可能なボイスの一覧を取得します（お気に入りが先頭に表示されます）
// @Tags voices
// @Accept json
// @Produce json
// @Param provider query string false "プロバイダでフィルタ（例: google）"
// @Param gender query string false "性別でフィルタ（male / female / neutral）"
// @Success 200 {object} response.VoiceListResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /voices [get]
func (h *VoiceHandler) ListVoices(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	var req request.ListVoicesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(err.Error()))
		return
	}

	filter := repository.VoiceFilter{
		Provider: req.Provider,
		Gender:   req.Gender,
	}

	voices, favIDs, err := h.voiceService.ListVoices(c.Request.Context(), userID, filter)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, http.StatusOK, toVoiceResponsesWithFavorites(voices, favIDs))
}

// GetVoice godoc
// @Summary ボイス取得
// @Description 指定された ID のボイスを取得します
// @Tags voices
// @Accept json
// @Produce json
// @Param voiceId path string true "ボイス ID（UUID 形式）"
// @Success 200 {object} response.VoiceDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /voices/{voiceId} [get]
func (h *VoiceHandler) GetVoice(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	id := c.Param("voiceId")

	// UUID 形式のバリデーション
	if err := uuid.Validate(id); err != nil {
		Error(c, err)
		return
	}

	voice, isFav, err := h.voiceService.GetVoice(c.Request.Context(), userID, id)
	if err != nil {
		Error(c, err)
		return
	}

	resp := toVoiceResponse(voice)
	resp.IsFavorite = isFav
	Success(c, http.StatusOK, resp)
}

// AddFavorite godoc
// @Summary ボイスお気に入り登録
// @Description 指定されたボイスをお気に入りに登録します
// @Tags voices
// @Accept json
// @Produce json
// @Param voiceId path string true "ボイス ID（UUID 形式）"
// @Success 201 {object} response.FavoriteVoiceResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /voices/{voiceId}/favorite [post]
func (h *VoiceHandler) AddFavorite(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	voiceID := c.Param("voiceId")
	if err := uuid.Validate(voiceID); err != nil {
		Error(c, err)
		return
	}

	fav, err := h.voiceService.AddFavorite(c.Request.Context(), userID, voiceID)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, http.StatusCreated, response.FavoriteVoiceResponse{
		VoiceID:   fav.VoiceID,
		CreatedAt: fav.CreatedAt,
	})
}

// RemoveFavorite godoc
// @Summary ボイスお気に入り解除
// @Description 指定されたボイスのお気に入りを解除します
// @Tags voices
// @Accept json
// @Produce json
// @Param voiceId path string true "ボイス ID（UUID 形式）"
// @Success 204 "No Content"
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /voices/{voiceId}/favorite [delete]
func (h *VoiceHandler) RemoveFavorite(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	voiceID := c.Param("voiceId")
	if err := uuid.Validate(voiceID); err != nil {
		Error(c, err)
		return
	}

	if err := h.voiceService.RemoveFavorite(c.Request.Context(), userID, voiceID); err != nil {
		Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// Voice モデルのスライスをレスポンス DTO のスライスに変換する（お気に入り情報付き）
func toVoiceResponsesWithFavorites(voices []model.Voice, favIDs []uuid.UUID) []response.VoiceResponse {
	result := make([]response.VoiceResponse, len(voices))
	for i, v := range voices {
		resp := toVoiceResponse(&v)
		resp.IsFavorite = containsUUID(favIDs, v.ID)
		result[i] = resp
	}
	return result
}

// Voice モデルをレスポンス DTO に変換する
func toVoiceResponse(v *model.Voice) response.VoiceResponse {
	return response.VoiceResponse{
		ID:              v.ID,
		Provider:        v.Provider,
		ProviderVoiceID: v.ProviderVoiceID,
		Name:            v.Name,
		Gender:          string(v.Gender),
		SampleAudioURL:  v.SampleAudioURL,
		IsActive:        v.IsActive,
	}
}

// containsUUID はスライスに指定された UUID が含まれるかを返す
func containsUUID(ids []uuid.UUID, id uuid.UUID) bool {
	for _, v := range ids {
		if v == id {
			return true
		}
	}
	return false
}
