package request

// お問い合わせ作成リクエスト
type CreateContactRequest struct {
	Category  string `json:"category" binding:"required,oneof=general bug_report feature_request other"`
	Email     string `json:"email" binding:"required,email"`
	Name      string `json:"name" binding:"required,min=1,max=100"`
	Content   string `json:"content" binding:"required,min=1,max=5000"`
	UserAgent string `json:"userAgent" binding:"omitempty"`
}
