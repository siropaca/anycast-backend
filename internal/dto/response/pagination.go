package response

// ページネーション情報のレスポンス（共通）
type PaginationResponse struct {
	Total  int64 `json:"total" validate:"required"`
	Limit  int   `json:"limit" validate:"required"`
	Offset int   `json:"offset" validate:"required"`
}
