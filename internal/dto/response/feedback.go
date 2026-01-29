package response

import (
	"time"

	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// フィードバック情報のレスポンス
type FeedbackResponse struct {
	ID         uuid.UUID        `json:"id" validate:"required"`
	Content    string           `json:"content" validate:"required"`
	Screenshot *ArtworkResponse `json:"screenshot" extensions:"x-nullable"`
	PageURL    *string          `json:"pageUrl" extensions:"x-nullable"`
	UserAgent  *string          `json:"userAgent" extensions:"x-nullable"`
	CreatedAt  time.Time        `json:"createdAt" validate:"required"`
}

// フィードバック単体のレスポンス
type FeedbackDataResponse struct {
	Data FeedbackResponse `json:"data" validate:"required"`
}
