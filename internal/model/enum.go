package model

// OAuthProvider は OAuth プロバイダの種別を表す
type OAuthProvider string

const (
	OAuthProviderGoogle OAuthProvider = "google"
)

// Gender は性別を表す
type Gender string

const (
	GenderMale    Gender = "male"
	GenderFemale  Gender = "female"
	GenderNeutral Gender = "neutral"
)

// Role はユーザーロールを表す
type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

// IsAdmin は管理者かどうかを判定する
func (r Role) IsAdmin() bool {
	return r == RoleAdmin
}
