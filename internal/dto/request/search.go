package request

// チャンネル検索リクエスト
type SearchChannelsRequest struct {
	PaginationRequest
	Q            string  `form:"q" binding:"required,max=255"`
	CategorySlug *string `form:"categorySlug" binding:"omitempty,max=50"`
}
