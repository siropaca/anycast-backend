package handler

import (
	"errors"

	"github.com/gin-gonic/gin"

	"github.com/siropaca/anycast-backend/internal/apperror"
)

// 成功レスポンスを返す
func Success(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"data": data})
}

// エラーレスポンスを返す
func Error(c *gin.Context, err error) {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		resp := gin.H{
			"code":    appErr.Code,
			"message": appErr.Message,
		}
		if appErr.Details != nil {
			resp["details"] = appErr.Details
		}
		c.JSON(appErr.HTTPStatus, gin.H{"error": resp})
		return
	}

	// 未知のエラーは 500
	c.JSON(500, gin.H{
		"error": gin.H{
			"code":    "INTERNAL_ERROR",
			"message": "Internal server error",
		},
	})
}
