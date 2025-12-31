package request

// ページネーションリクエスト（共通）
type PaginationRequest struct {
	Limit  int `form:"limit,default=20" binding:"min=1,max=100"`
	Offset int `form:"offset,default=0" binding:"min=0"`
}
