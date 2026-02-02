package request

// おすすめチャンネル取得リクエスト
type RecommendChannelsRequest struct {
	PaginationRequest
	CategorySlug *string `form:"categorySlug" binding:"omitempty,max=50"`
}

// おすすめエピソード取得リクエスト
type RecommendEpisodesRequest struct {
	PaginationRequest
	CategorySlug *string `form:"categorySlug" binding:"omitempty,max=50"`
}
