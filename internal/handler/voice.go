package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/infrastructure/storage"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
	"github.com/siropaca/anycast-backend/internal/repository"
	"github.com/siropaca/anycast-backend/internal/service"
)

// ボイス関連のハンドラー
type VoiceHandler struct {
	voiceService  service.VoiceService
	storageClient storage.Client
}

// VoiceHandler を作成する
func NewVoiceHandler(vs service.VoiceService, sc storage.Client) *VoiceHandler {
	return &VoiceHandler{voiceService: vs, storageClient: sc}
}

// ListVoices godoc
// @Summary ボイス一覧取得
// @Description 利用可能なボイスの一覧を取得します
// @Tags voices
// @Accept json
// @Produce json
// @Param provider query string false "プロバイダでフィルタ（例: google）"
// @Param gender query string false "性別でフィルタ（male / female / neutral）"
// @Success 200 {object} response.VoiceListResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
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

	Success(c, http.StatusOK, toVoiceResponses(voices, h.storageClient, c))
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
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /voices/{voiceId} [get]
func (h *VoiceHandler) GetVoice(c *gin.Context) {
	id := c.Param("voiceId")

	// UUID 形式のバリデーション
	if err := uuid.Validate(id); err != nil {
		Error(c, err)
		return
	}

	voice, err := h.voiceService.GetVoice(c.Request.Context(), id)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, http.StatusOK, toVoiceResponse(voice, h.storageClient, c))
}

// Voice モデルのスライスをレスポンス DTO のスライスに変換する
func toVoiceResponses(voices []model.Voice, storageClient storage.Client, c *gin.Context) []response.VoiceResponse {
	result := make([]response.VoiceResponse, len(voices))
	for i, v := range voices {
		result[i] = toVoiceResponse(&v, storageClient, c)
	}
	return result
}

// Voice モデルをレスポンス DTO に変換する
func toVoiceResponse(v *model.Voice, storageClient storage.Client, c *gin.Context) response.VoiceResponse {
	sampleAudioURL := ""
	if storageClient != nil {
		signedURL, err := storageClient.GenerateSignedURL(c.Request.Context(), v.SampleAudio.Path, 1*time.Hour)
		if err == nil {
			sampleAudioURL = signedURL
		}
	}

	return response.VoiceResponse{
		ID:              v.ID,
		Provider:        v.Provider,
		ProviderVoiceID: v.ProviderVoiceID,
		Name:            v.Name,
		Gender:          string(v.Gender),
		SampleAudioURL:  sampleAudioURL,
		IsActive:        v.IsActive,
	}
}
