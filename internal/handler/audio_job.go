package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/middleware"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/repository"
	"github.com/siropaca/anycast-backend/internal/service"
)

// AudioJobHandler は音声生成ジョブ関連のハンドラー
type AudioJobHandler struct {
	audioJobService service.AudioJobService
}

// NewAudioJobHandler は AudioJobHandler を作成する
func NewAudioJobHandler(ajs service.AudioJobService) *AudioJobHandler {
	return &AudioJobHandler{audioJobService: ajs}
}

// GenerateAudioAsync godoc
// @Summary 非同期音声生成
// @Description エピソードの音声を非同期で生成します。ジョブを作成し、完了時は WebSocket で通知されます。
// @Tags episodes
// @Accept json
// @Produce json
// @Param channelId path string true "チャンネル ID"
// @Param episodeId path string true "エピソード ID"
// @Param body body request.GenerateAudioAsyncRequest false "音声生成オプション"
// @Success 202 {object} response.AudioJobDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /channels/{channelId}/episodes/{episodeId}/audio/generate-async [post]
func (h *AudioJobHandler) GenerateAudioAsync(c *gin.Context) {
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

	var req request.GenerateAudioAsyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// JSON が空でもエラーにしない
		req = request.GenerateAudioAsyncRequest{}
	}

	result, err := h.audioJobService.CreateJob(c.Request.Context(), userID, channelID, episodeID, req)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"data": result})
}

// GetAudioJob godoc
// @Summary 音声生成ジョブ詳細取得
// @Description 音声生成ジョブの詳細を取得します
// @Tags audio-jobs
// @Accept json
// @Produce json
// @Param jobId path string true "ジョブ ID"
// @Success 200 {object} response.AudioJobDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /audio-jobs/{jobId} [get]
func (h *AudioJobHandler) GetAudioJob(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	jobID := c.Param("jobId")
	if jobID == "" {
		Error(c, apperror.ErrValidation.WithMessage("jobId は必須です"))
		return
	}

	result, err := h.audioJobService.GetJob(c.Request.Context(), userID, jobID)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// ListMyAudioJobs godoc
// @Summary 自分の音声生成ジョブ一覧取得
// @Description 自分の音声生成ジョブ一覧を取得します
// @Tags me
// @Accept json
// @Produce json
// @Param status query string false "ステータスでフィルタ（pending / processing / completed / failed）"
// @Success 200 {object} response.AudioJobListResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /me/audio-jobs [get]
func (h *AudioJobHandler) ListMyAudioJobs(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	var req request.ListMyAudioJobsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(err.Error()))
		return
	}

	filter := repository.AudioJobFilter{}

	// ステータスフィルタ
	if req.Status != nil {
		status := model.AudioJobStatus(*req.Status)
		filter.Status = &status
	}

	result, err := h.audioJobService.ListMyJobs(c.Request.Context(), userID, filter)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}
