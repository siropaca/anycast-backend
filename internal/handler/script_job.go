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

// ScriptJobHandler は台本生成ジョブ関連のハンドラー
type ScriptJobHandler struct {
	scriptJobService service.ScriptJobService
}

// NewScriptJobHandler は ScriptJobHandler を作成する
func NewScriptJobHandler(sjs service.ScriptJobService) *ScriptJobHandler {
	return &ScriptJobHandler{scriptJobService: sjs}
}

// GenerateScriptAsync godoc
// @Summary 非同期台本生成
// @Description エピソードの台本を非同期で生成します。ジョブを作成し、完了時は WebSocket で通知されます。
// @Tags script
// @Accept json
// @Produce json
// @Param channelId path string true "チャンネル ID"
// @Param episodeId path string true "エピソード ID"
// @Param body body request.GenerateScriptAsyncRequest true "台本生成オプション"
// @Success 202 {object} response.ScriptJobDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /channels/{channelId}/episodes/{episodeId}/script/generate-async [post]
func (h *ScriptJobHandler) GenerateScriptAsync(c *gin.Context) {
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

	var req request.GenerateScriptAsyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(formatValidationError(err)))
		return
	}

	result, err := h.scriptJobService.CreateJob(c.Request.Context(), userID, channelID, episodeID, req)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"data": result})
}

// GetScriptJob godoc
// @Summary 台本生成ジョブ詳細取得
// @Description 台本生成ジョブの詳細を取得します
// @Tags script-jobs
// @Accept json
// @Produce json
// @Param jobId path string true "ジョブ ID"
// @Success 200 {object} response.ScriptJobDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /script-jobs/{jobId} [get]
func (h *ScriptJobHandler) GetScriptJob(c *gin.Context) {
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

	result, err := h.scriptJobService.GetJob(c.Request.Context(), userID, jobID)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// ListMyScriptJobs godoc
// @Summary 自分の台本生成ジョブ一覧取得
// @Description 自分の台本生成ジョブ一覧を取得します
// @Tags me
// @Accept json
// @Produce json
// @Param status query string false "ステータスでフィルタ（pending / processing / completed / failed）"
// @Success 200 {object} response.ScriptJobListResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /me/script-jobs [get]
func (h *ScriptJobHandler) ListMyScriptJobs(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		Error(c, apperror.ErrUnauthorized)
		return
	}

	var req request.ListMyScriptJobsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(formatValidationError(err)))
		return
	}

	filter := repository.ScriptJobFilter{}

	// ステータスフィルタ
	if req.Status != nil {
		status := model.ScriptJobStatus(*req.Status)
		filter.Status = &status
	}

	result, err := h.scriptJobService.ListMyJobs(c.Request.Context(), userID, filter)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetLatestScriptJob godoc
// @Summary 最新の完了済み台本生成ジョブ取得
// @Description エピソードの最新の完了済み台本生成ジョブを取得します。完了済みジョブが存在しない場合は data: null を返します。
// @Tags script-jobs
// @Accept json
// @Produce json
// @Param channelId path string true "チャンネル ID"
// @Param episodeId path string true "エピソード ID"
// @Success 200 {object} response.ScriptJobDataNullableResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /channels/{channelId}/episodes/{episodeId}/script-jobs/latest [get]
func (h *ScriptJobHandler) GetLatestScriptJob(c *gin.Context) {
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

	result, err := h.scriptJobService.GetLatestJobByEpisode(c.Request.Context(), userID, channelID, episodeID)
	if err != nil {
		Error(c, err)
		return
	}

	if result == nil {
		c.JSON(http.StatusOK, gin.H{"data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// GenerateScriptDirect godoc
// @Summary 開発用: 台本直接生成
// @Description DB を使わずにリクエストパラメータのみで台本を生成します（開発・検証用）
// @Tags dev
// @Accept json
// @Produce json
// @Param body body request.GenerateScriptDirectRequest true "台本生成パラメータ"
// @Success 200 {object} response.GenerateScriptDirectResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /dev/script/generate [post]
func (h *ScriptJobHandler) GenerateScriptDirect(c *gin.Context) {
	var req request.GenerateScriptDirectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage(formatValidationError(err)))
		return
	}

	result, err := h.scriptJobService.GenerateScriptDirect(c.Request.Context(), req)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// CancelScriptJob godoc
// @Summary 台本生成ジョブキャンセル
// @Description 台本生成ジョブをキャンセルします。pending 状態のジョブは即座に canceled に、processing 状態のジョブは canceling に遷移します。
// @Tags script-jobs
// @Accept json
// @Produce json
// @Param jobId path string true "ジョブ ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /script-jobs/{jobId}/cancel [post]
func (h *ScriptJobHandler) CancelScriptJob(c *gin.Context) {
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

	if err := h.scriptJobService.CancelJob(c.Request.Context(), userID, jobID); err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
