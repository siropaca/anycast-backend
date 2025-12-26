package request

// ListVoicesRequest は GET /voices のクエリパラメータ
type ListVoicesRequest struct {
	Provider *string `form:"provider"`
	Gender   *string `form:"gender"`
}
