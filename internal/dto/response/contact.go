package response

import (
	"time"

	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// お問い合わせ情報のレスポンス
type ContactResponse struct {
	ID        uuid.UUID `json:"id" validate:"required"`
	Category  string    `json:"category" validate:"required"`
	Email     string    `json:"email" validate:"required"`
	Name      string    `json:"name" validate:"required"`
	Content   string    `json:"content" validate:"required"`
	UserAgent *string   `json:"userAgent" extensions:"x-nullable"`
	CreatedAt time.Time `json:"createdAt" validate:"required"`
}

// お問い合わせ単体のレスポンス
type ContactDataResponse struct {
	Data ContactResponse `json:"data" validate:"required"`
}
