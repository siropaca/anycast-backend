package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/service"
)

// フィードバック関連のハンドラー
type FeedbackHandler struct {
	feedbackService service.FeedbackService
}

// FeedbackHandler を作成する
func NewFeedbackHandler(fs service.FeedbackService) *FeedbackHandler {
	return &FeedbackHandler{feedbackService: fs}
}

// CreateFeedback godoc
// @Summary フィードバック送信
// @Description フィードバックを送信します
// @Tags feedbacks
// @Accept multipart/form-data
// @Produce json
// @Param content formData string true "フィードバック内容（1〜5000文字）"
// @Param screenshot formData file false "スクリーンショット画像（png, jpeg, webp）"
// @Param pageUrl formData string false "現在のページ URL"
// @Param userAgent formData string false "ブラウザの User-Agent"
// @Success 201 {object} response.FeedbackDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /feedbacks [post]
func (h *FeedbackHandler) CreateFeedback(c *gin.Context) {
	userID := c.GetString("userID")

	// フィードバック内容を取得
	content := c.PostForm("content")
	if content == "" {
		Error(c, apperror.ErrValidation.WithMessage("フィードバック内容は必須です"))
		return
	}
	if len(content) > 5000 {
		Error(c, apperror.ErrValidation.WithMessage("フィードバック内容は5000文字以内で入力してください"))
		return
	}

	input := service.CreateFeedbackInput{
		Content: content,
	}

	// ページ URL（任意）
	if pageURL := c.PostForm("pageUrl"); pageURL != "" {
		input.PageURL = &pageURL
	}

	// User-Agent（任意）
	if userAgent := c.PostForm("userAgent"); userAgent != "" {
		input.UserAgent = &userAgent
	}

	// スクリーンショット（任意）
	if fileHeader, err := c.FormFile("screenshot"); err == nil {
		file, err := fileHeader.Open()
		if err != nil {
			Error(c, apperror.ErrInternal.WithMessage("ファイルを開けませんでした").WithError(err))
			return
		}
		defer file.Close()

		// MIME タイプのバリデーション
		contentType := fileHeader.Header.Get("Content-Type")
		if !isValidImageType(contentType) {
			Error(c, apperror.ErrValidation.WithMessage("スクリーンショットは png, jpeg, webp 形式のみ対応しています"))
			return
		}

		input.Screenshot = &service.UploadImageInput{
			File:        file,
			Filename:    fileHeader.Filename,
			ContentType: contentType,
			FileSize:    int(fileHeader.Size),
		}
	}

	result, err := h.feedbackService.CreateFeedback(c.Request.Context(), userID, input)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusCreated, result)
}

// isValidImageType は許可された画像形式かチェックする
func isValidImageType(contentType string) bool {
	allowedTypes := map[string]bool{
		"image/png":  true,
		"image/jpeg": true,
		"image/webp": true,
	}
	return allowedTypes[contentType]
}
