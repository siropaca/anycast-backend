package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/siropaca/anycast-backend/internal/apperror"
	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/middleware"
	"github.com/siropaca/anycast-backend/internal/service"
)

// お問い合わせ関連のハンドラー
type ContactHandler struct {
	contactService service.ContactService
}

// ContactHandler を作成する
func NewContactHandler(cs service.ContactService) *ContactHandler {
	return &ContactHandler{contactService: cs}
}

// CreateContact godoc
// @Summary お問い合わせ送信
// @Description お問い合わせを送信します（認証任意）
// @Tags contacts
// @Accept json
// @Produce json
// @Param body body request.CreateContactRequest true "お問い合わせ内容"
// @Success 201 {object} response.ContactDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /contacts [post]
func (h *ContactHandler) CreateContact(c *gin.Context) {
	var req request.CreateContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, apperror.ErrValidation.WithMessage("リクエストの形式が正しくありません").WithError(err))
		return
	}

	input := service.CreateContactInput{
		Category: req.Category,
		Email:    req.Email,
		Name:     req.Name,
		Content:  req.Content,
	}

	// User-Agent（任意）
	if req.UserAgent != "" {
		input.UserAgent = &req.UserAgent
	}

	// 任意認証: ログイン済みの場合は userID を取得
	if userID, ok := middleware.GetUserID(c); ok {
		input.UserID = &userID
	}

	result, err := h.contactService.CreateContact(c.Request.Context(), input)
	if err != nil {
		Error(c, err)
		return
	}

	c.JSON(http.StatusCreated, result)
}
