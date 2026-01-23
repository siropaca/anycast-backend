package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
	"github.com/siropaca/anycast-backend/internal/service"
)

// WorkerHandler は Cloud Tasks ワーカー用のハンドラー
type WorkerHandler struct {
	audioJobService  service.AudioJobService
	scriptJobService service.ScriptJobService
}

// NewWorkerHandler は WorkerHandler を作成する
func NewWorkerHandler(ajs service.AudioJobService, sjs service.ScriptJobService) *WorkerHandler {
	return &WorkerHandler{
		audioJobService:  ajs,
		scriptJobService: sjs,
	}
}

// AudioJobPayload は音声ワーカーに送信されるペイロード
type AudioJobPayload struct {
	JobID string `json:"jobId" binding:"required"`
}

// ScriptJobPayload は台本ワーカーに送信されるペイロード
type ScriptJobPayload struct {
	JobID string `json:"jobId" binding:"required"`
}

// ProcessAudioJob godoc
// @Summary 音声生成ジョブを処理
// @Description Cloud Tasks から呼び出される音声生成ワーカーエンドポイント
// @Tags internal
// @Accept json
// @Produce json
// @Param payload body AudioJobPayload true "ジョブ情報"
// @Success 200 {object} map[string]string
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /internal/worker/audio [post]
func (h *WorkerHandler) ProcessAudioJob(c *gin.Context) {
	log := logger.FromContext(c.Request.Context())

	var payload AudioJobPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		log.Error("ペイロードが不正です", "error", err)
		Error(c, apperror.ErrValidation.WithMessage("jobId は必須です"))
		return
	}

	log.Info("音声ジョブを処理中", "job_id", payload.JobID)

	if err := h.audioJobService.ExecuteJob(c.Request.Context(), payload.JobID); err != nil {
		log.Error("ジョブの実行に失敗しました", "error", err, "job_id", payload.JobID)
		// Cloud Tasks はエラーレスポンスを受け取るとリトライするため、
		// ビジネスエラーでも 200 を返す（ジョブ自体は失敗状態で記録される）
		// 500 を返すのはリトライ可能なエラーのみ
		if apperror.IsRetryable(err) {
			Error(c, err)
			return
		}
		// 非リトライエラーは 200 で返す（ジョブは失敗状態）
		c.JSON(http.StatusOK, gin.H{
			"status":  "failed",
			"job_id":  payload.JobID,
			"message": "job failed but should not retry",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "completed",
		"job_id": payload.JobID,
	})
}

// ProcessScriptJob godoc
// @Summary 台本生成ジョブを処理
// @Description Cloud Tasks から呼び出される台本生成ワーカーエンドポイント
// @Tags internal
// @Accept json
// @Produce json
// @Param payload body ScriptJobPayload true "ジョブ情報"
// @Success 200 {object} map[string]string
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /internal/worker/script [post]
func (h *WorkerHandler) ProcessScriptJob(c *gin.Context) {
	log := logger.FromContext(c.Request.Context())

	var payload ScriptJobPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		log.Error("ペイロードが不正です", "error", err)
		Error(c, apperror.ErrValidation.WithMessage("jobId は必須です"))
		return
	}

	log.Info("台本ジョブを処理中", "job_id", payload.JobID)

	if err := h.scriptJobService.ExecuteJob(c.Request.Context(), payload.JobID); err != nil {
		log.Error("台本ジョブの実行に失敗しました", "error", err, "job_id", payload.JobID)
		// Cloud Tasks はエラーレスポンスを受け取るとリトライするため、
		// ビジネスエラーでも 200 を返す（ジョブ自体は失敗状態で記録される）
		// 500 を返すのはリトライ可能なエラーのみ
		if apperror.IsRetryable(err) {
			Error(c, err)
			return
		}
		// 非リトライエラーは 200 で返す（ジョブは失敗状態）
		c.JSON(http.StatusOK, gin.H{
			"status":  "failed",
			"job_id":  payload.JobID,
			"message": "job failed but should not retry",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "completed",
		"job_id": payload.JobID,
	})
}
