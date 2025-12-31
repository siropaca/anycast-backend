package response

import "github.com/google/uuid"

// アートワーク情報のレスポンス
type ArtworkResponse struct {
	ID  uuid.UUID `json:"id"`
	URL string    `json:"url"`
}
