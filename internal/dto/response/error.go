package response

// エラーレスポンスの詳細
type ErrorDetail struct {
	Code    string `json:"code" example:"VALIDATION_ERROR"`
	Message string `json:"message" example:"Invalid request parameters"`
	Details any    `json:"details,omitempty"`
}

// エラーレスポンス
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}
