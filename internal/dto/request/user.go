package request

// ユーザー情報更新リクエスト
type UpdateMeRequest struct {
	DisplayName   string  `json:"displayName" binding:"required,max=20"`
	Bio           string  `json:"bio" binding:"max=200"`
	AvatarImageID *string `json:"avatarImageId" binding:"omitempty,uuid"`
	HeaderImageID *string `json:"headerImageId" binding:"omitempty,uuid"`
}
