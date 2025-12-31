package response

import "github.com/google/uuid"

// カテゴリ情報のレスポンス
type CategoryResponse struct {
	ID   uuid.UUID `json:"id" validate:"required"`
	Slug string    `json:"slug" validate:"required"`
	Name string    `json:"name" validate:"required"`
}
