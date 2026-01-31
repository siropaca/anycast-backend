package model

// OAuthProvider は OAuth プロバイダのタイプを表す
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

// ReactionType はリアクションのタイプを表す
type ReactionType string

const (
	ReactionTypeLike ReactionType = "like"
	ReactionTypeBad  ReactionType = "bad"
)
