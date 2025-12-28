package response

import "github.com/google/uuid"

// ユーザー情報のレスポンス
type UserResponse struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	Username    string    `json:"username"`
	DisplayName string    `json:"displayName"`
	AvatarURL   *string   `json:"avatarUrl"`
}

// ユーザー単体のレスポンス
type UserDataResponse struct {
	Data UserResponse `json:"data"`
}
