package response

import (
	"time"

	"github.com/siropaca/anycast-backend/internal/pkg/uuid"
)

// フォロー中のユーザー情報のレスポンス
type FollowUserResponse struct {
	ID          uuid.UUID       `json:"id" validate:"required"`
	Username    string          `json:"username" validate:"required"`
	DisplayName string          `json:"displayName" validate:"required"`
	Avatar      *AvatarResponse `json:"avatar" extensions:"x-nullable"`
}

// フォローアイテムのレスポンス
type FollowItemResponse struct {
	User       FollowUserResponse `json:"user" validate:"required"`
	FollowedAt time.Time          `json:"followedAt" validate:"required"`
}

// フォロー中のユーザー一覧（ページネーション付き）のレスポンス
type FollowListWithPaginationResponse struct {
	Data       []FollowItemResponse `json:"data" validate:"required"`
	Pagination PaginationResponse   `json:"pagination" validate:"required"`
}
