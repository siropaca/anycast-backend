package response

// エラーレスポンスの詳細
type ErrorDetail struct {
	Code    string `json:"code" validate:"required" example:"VALIDATION_ERROR"`
	Message string `json:"message" validate:"required" example:"Invalid request parameters"`
	Details any    `json:"details,omitempty"`
}

// エラーレスポンス
type ErrorResponse struct {
	Error ErrorDetail `json:"error" validate:"required"`
}
