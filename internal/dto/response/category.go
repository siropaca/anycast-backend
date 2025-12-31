package response

import "github.com/google/uuid"

// カテゴリ情報のレスポンス
type CategoryResponse struct {
	ID   uuid.UUID `json:"id"`
	Slug string    `json:"slug"`
	Name string    `json:"name"`
}
