package response

import (
	"time"

	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// API キー一覧用レスポンス
type APIKeyResponse struct {
	ID         uuid.UUID  `json:"id" validate:"required"`
	Name       string     `json:"name" validate:"required"`
	Prefix     string     `json:"prefix" validate:"required"`
	LastUsedAt *time.Time `json:"lastUsedAt" extensions:"x-nullable"`
	CreatedAt  time.Time  `json:"createdAt" validate:"required"`
}

// API キー一覧レスポンス（data ラッパー）
type APIKeyListDataResponse struct {
	Data []APIKeyResponse `json:"data" validate:"required"`
}

// API キー作成時レスポンス（平文キー含む）
type APIKeyCreatedResponse struct {
	ID        uuid.UUID `json:"id" validate:"required"`
	Name      string    `json:"name" validate:"required"`
	Key       string    `json:"key" validate:"required"`
	Prefix    string    `json:"prefix" validate:"required"`
	CreatedAt time.Time `json:"createdAt" validate:"required"`
}

// API キー作成時レスポンス（data ラッパー）
type APIKeyCreatedDataResponse struct {
	Data APIKeyCreatedResponse `json:"data" validate:"required"`
}
