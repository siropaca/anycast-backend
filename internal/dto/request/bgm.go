package request

// 自分の BGM 一覧取得リクエスト
type ListMyBgmsRequest struct {
	PaginationRequest
	IncludeDefault bool `form:"include_default,default=false"`
}
