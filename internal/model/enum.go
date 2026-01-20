package model

// OAuth プロバイダの種別
type OAuthProvider string

const (
	OAuthProviderGoogle OAuthProvider = "google"
)

// 性別
type Gender string

const (
	GenderMale    Gender = "male"
	GenderFemale  Gender = "female"
	GenderNeutral Gender = "neutral"
)

// ユーザーロール
type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

// 管理者かどうかを判定
func (r Role) IsAdmin() bool {
	return r == RoleAdmin
}
