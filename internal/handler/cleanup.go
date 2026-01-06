package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/infrastructure/storage"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/service"
)

// クリーンアップ関連のハンドラー
type CleanupHandler struct {
	cleanupService service.CleanupService
	storageClient  storage.Client
}

// CleanupHandler を作成する
func NewCleanupHandler(cs service.CleanupService, sc storage.Client) *CleanupHandler {
	return &CleanupHandler{cleanupService: cs, storageClient: sc}
}

// CleanupOrphanedMedia godoc
// @Summary 孤児メディアファイル削除
// @Description どのテーブルからも参照されていない audios / images レコードを検出し、GCS ファイルと DB レコードを削除する
// @Tags admin
// @Accept json
// @Produce json
// @Param dry_run query bool false "true の場合、削除対象の一覧を返すのみで実際の削除は行わない"
// @Success 200 {object} response.CleanupOrphanedMediaResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/cleanup/orphaned-media [post]
func (h *CleanupHandler) CleanupOrphanedMedia(c *gin.Context) {
	dryRun := c.DefaultQuery("dry_run", "false") == "true"

	result, err := h.cleanupService.CleanupOrphanedMedia(c.Request.Context(), dryRun)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, http.StatusOK, toCleanupResponse(result, dryRun, h.storageClient, c))
}

// CleanupResult をレスポンス DTO に変換する
func toCleanupResponse(result *service.CleanupResult, dryRun bool, storageClient storage.Client, c *gin.Context) response.CleanupOrphanedMediaResponse {
	return response.CleanupOrphanedMediaResponse{
		DryRun:            dryRun,
		OrphanedAudios:    toOrphanedAudioResponses(result.OrphanedAudios, storageClient, c),
		OrphanedImages:    toOrphanedImageResponses(result.OrphanedImages, storageClient, c),
		DeletedAudioCount: result.DeletedAudioCount,
		DeletedImageCount: result.DeletedImageCount,
		FailedAudioCount:  result.FailedAudioCount,
		FailedImageCount:  result.FailedImageCount,
	}
}

// Audio モデルのスライスをレスポンス DTO のスライスに変換する
func toOrphanedAudioResponses(audios []model.Audio, storageClient storage.Client, c *gin.Context) []response.OrphanedAudioResponse {
	result := make([]response.OrphanedAudioResponse, len(audios))
	for i, audio := range audios {
		result[i] = toOrphanedAudioResponse(&audio, storageClient, c)
	}
	return result
}

// Audio モデルをレスポンス DTO に変換する
func toOrphanedAudioResponse(audio *model.Audio, storageClient storage.Client, c *gin.Context) response.OrphanedAudioResponse {
	url := ""
	if storageClient != nil {
		signedURL, err := storageClient.GenerateSignedURL(c.Request.Context(), audio.Path, 1*time.Hour)
		if err == nil {
			url = signedURL
		}
	}

	return response.OrphanedAudioResponse{
		ID:        audio.ID,
		URL:       url,
		Filename:  audio.Filename,
		FileSize:  audio.FileSize,
		CreatedAt: audio.CreatedAt,
	}
}

// Image モデルのスライスをレスポンス DTO のスライスに変換する
func toOrphanedImageResponses(images []model.Image, storageClient storage.Client, c *gin.Context) []response.OrphanedImageResponse {
	result := make([]response.OrphanedImageResponse, len(images))
	for i, image := range images {
		result[i] = toOrphanedImageResponse(&image, storageClient, c)
	}
	return result
}

// Image モデルをレスポンス DTO に変換する
func toOrphanedImageResponse(image *model.Image, storageClient storage.Client, c *gin.Context) response.OrphanedImageResponse {
	url := ""
	if storageClient != nil {
		signedURL, err := storageClient.GenerateSignedURL(c.Request.Context(), image.URL, 1*time.Hour)
		if err == nil {
			url = signedURL
		}
	}

	return response.OrphanedImageResponse{
		ID:        image.ID,
		URL:       url,
		Filename:  image.Filename,
		FileSize:  image.FileSize,
		CreatedAt: image.CreatedAt,
	}
}
