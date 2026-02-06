package request

// ユーザー登録リクエスト
type RegisterRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8,max=100"`
	DisplayName string `json:"displayName" binding:"required,max=20"`
}

// ログインリクエスト
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Google OAuth 認証リクエスト
type OAuthGoogleRequest struct {
	ProviderUserID string  `json:"providerUserId" binding:"required"`
	Email          string  `json:"email" binding:"required,email"`
	DisplayName    string  `json:"displayName" binding:"required,max=20"`
	AccessToken    string  `json:"accessToken" binding:"required"`
	RefreshToken   *string `json:"refreshToken"`
	ExpiresAt      *int64  `json:"expiresAt"`
}

// トークンリフレッシュリクエスト
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// ログアウトリクエスト
type LogoutRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// ユーザープロンプト更新リクエスト
type UpdateUserPromptRequest struct {
	UserPrompt *string `json:"userPrompt" binding:"omitempty,max=2000"`
}
