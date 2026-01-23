package request

// ボイス一覧取得リクエスト
type ListVoicesRequest struct {
	Provider *string `form:"provider"`
	Gender   *string `form:"gender"`
}
