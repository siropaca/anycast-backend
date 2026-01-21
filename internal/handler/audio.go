package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/service"
)

// 音声関連のハンドラー
type AudioHandler struct {
	audioService service.AudioService
}

// AudioHandler を作成する
func NewAudioHandler(as service.AudioService) *AudioHandler {
	return &AudioHandler{audioService: as}
}

// UploadAudio godoc
// @Summary 音声アップロード
// @Description 音声ファイルをアップロードします
// @Tags audios
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "アップロードする音声ファイル（mp3, wav, ogg, aac, m4a）"
// @Success 201 {object} response.AudioUploadDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /audios [post]
func (h *AudioHandler) UploadAudio(c *gin.Context) {
	// ファイルの取得
	fileHeader, err := c.FormFile("file")
	if err != nil {
		Error(c, apperror.ErrValidation.WithMessage("ファイルは必須です"))
		return
	}

	// ファイルを開く
	file, err := fileHeader.Open()
	if err != nil {
		Error(c, apperror.ErrInternal.WithMessage("ファイルを開けませんでした").WithError(err))
		return
	}
	defer file.Close()

	// サービスに渡す入力データを作成
	input := service.UploadAudioInput{
		File:        file,
		Filename:    fileHeader.Filename,
		ContentType: fileHeader.Header.Get("Content-Type"),
		FileSize:    int(fileHeader.Size),
	}

	result, err := h.audioService.UploadAudio(c.Request.Context(), input)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusCreated, result)
}
