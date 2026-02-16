package request

import "github.com/siropaca/anycast-backend/internal/pkg/optional"

// ユーザー情報更新リクエスト
type UpdateMeRequest struct {
	DisplayName   string                 `json:"displayName" binding:"required,max=20"`
	Bio           string                 `json:"bio" binding:"max=200"`
	AvatarImageID optional.Field[string] `json:"avatarImageId"`
	HeaderImageID optional.Field[string] `json:"headerImageId"`
}
