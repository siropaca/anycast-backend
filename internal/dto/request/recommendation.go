package request

// おすすめチャンネル取得リクエスト
type RecommendChannelsRequest struct {
	PaginationRequest
	CategoryID *string `form:"categoryId" binding:"omitempty,uuid"`
}
