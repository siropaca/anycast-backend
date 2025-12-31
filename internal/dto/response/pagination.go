package response

// ページネーション情報のレスポンス（共通）
type PaginationResponse struct {
	Total  int64 `json:"total"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
}
