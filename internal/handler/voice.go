package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/repository"
	"github.com/siropaca/anycast-backend/internal/service"
)

// VoiceHandler はボイス関連のハンドラー
type VoiceHandler struct {
	voiceService service.VoiceService
}

// NewVoiceHandler は VoiceHandler を作成する
func NewVoiceHandler(vs service.VoiceService) *VoiceHandler {
	return &VoiceHandler{voiceService: vs}
}

// ListVoices godoc
// @Summary ボイス一覧取得
// @Description 利用可能なボイスの一覧を取得します
// @Tags voices
// @Accept json
// @Produce json
// @Param provider query string false "プロバイダでフィルタ（例: google）"
// @Param gender query string false "性別でフィルタ（male / female / neutral）"
// @Success 200 {object} map[string][]response.VoiceResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /voices [get]
func (h *VoiceHandler) ListVoices(c *gin.Context) {
	var req request.ListVoicesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(err.Error()))
		return
	}

	filter := repository.VoiceFilter{
		Provider: req.Provider,
		Gender:   req.Gender,
	}

	voices, err := h.voiceService.ListVoices(c.Request.Context(), filter)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, http.StatusOK, toVoiceResponses(voices))
}

// GetVoice godoc
// @Summary ボイス取得
// @Description 指定された ID のボイスを取得します
// @Tags voices
// @Accept json
// @Produce json
// @Param voiceId path string true "ボイス ID（UUID 形式）"
// @Success 200 {object} map[string]response.VoiceResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /voices/{voiceId} [get]
func (h *VoiceHandler) GetVoice(c *gin.Context) {
	id := c.Param("voiceId")

	// UUID 形式のバリデーション
	if _, err := uuid.Parse(id); err != nil {
		Error(c, apperror.ErrValidation.WithMessage("Invalid voice ID format"))
		return
	}

	voice, err := h.voiceService.GetVoice(c.Request.Context(), id)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, http.StatusOK, toVoiceResponse(voice))
}

// toVoiceResponses は Voice モデルのスライスをレスポンス DTO のスライスに変換する
func toVoiceResponses(voices []model.Voice) []response.VoiceResponse {
	result := make([]response.VoiceResponse, len(voices))
	for i, v := range voices {
		result[i] = toVoiceResponse(&v)
	}
	return result
}

// toVoiceResponse は Voice モデルをレスポンス DTO に変換する
func toVoiceResponse(v *model.Voice) response.VoiceResponse {
	return response.VoiceResponse{
		ID:              v.ID,
		Provider:        v.Provider,
		ProviderVoiceID: v.ProviderVoiceID,
		Name:            v.Name,
		Gender:          v.Gender,
		IsActive:        v.IsActive,
	}
}
