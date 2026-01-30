package request

// おすすめチャンネル取得リクエスト
type RecommendChannelsRequest struct {
	PaginationRequest
	CategoryID *string `form:"categoryId" binding:"omitempty,uuid"`
}

// おすすめエピソード取得リクエスト
type RecommendEpisodesRequest struct {
	PaginationRequest
	CategoryID *string `form:"categoryId" binding:"omitempty,uuid"`
}
