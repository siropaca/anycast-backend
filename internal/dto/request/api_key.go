package request

// API キー作成リクエスト
type CreateAPIKeyRequest struct {
	Name string `json:"name" binding:"required,min=1,max=100"`
}
