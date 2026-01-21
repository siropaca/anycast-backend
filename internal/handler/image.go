package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/service"
)

// 画像関連のハンドラー
type ImageHandler struct {
	imageService service.ImageService
}

// ImageHandler を作成する
func NewImageHandler(is service.ImageService) *ImageHandler {
	return &ImageHandler{imageService: is}
}

// UploadImage godoc
// @Summary 画像アップロード
// @Description 画像ファイルをアップロードします
// @Tags images
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "アップロードする画像ファイル（png, jpeg, gif, webp）"
// @Success 201 {object} response.ImageUploadDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /images [post]
func (h *ImageHandler) UploadImage(c *gin.Context) {
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
	input := service.UploadImageInput{
		File:        file,
		Filename:    fileHeader.Filename,
		ContentType: fileHeader.Header.Get("Content-Type"),
		FileSize:    int(fileHeader.Size),
	}

	result, err := h.imageService.UploadImage(c.Request.Context(), input)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusCreated, result)
}
