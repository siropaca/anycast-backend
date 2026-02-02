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

// ContactCategory はお問い合わせカテゴリを表す
type ContactCategory string

const (
	ContactCategoryGeneral        ContactCategory = "general"
	ContactCategoryBugReport      ContactCategory = "bug_report"
	ContactCategoryFeatureRequest ContactCategory = "feature_request"
	ContactCategoryOther          ContactCategory = "other"
)

// Label はカテゴリの日本語ラベルを返す
func (c ContactCategory) Label() string {
	switch c {
	case ContactCategoryGeneral:
		return "一般的なお問い合わせ"
	case ContactCategoryBugReport:
		return "不具合の報告"
	case ContactCategoryFeatureRequest:
		return "機能リクエスト"
	case ContactCategoryOther:
		return "その他"
	default:
		return string(c)
	}
}
