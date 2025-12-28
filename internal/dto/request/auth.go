package request

// POST /auth/register のリクエストボディ
type RegisterRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8,max=100"`
	DisplayName string `json:"displayName" binding:"required,max=20"`
}

// POST /auth/login のリクエストボディ
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// POST /auth/oauth/google のリクエストボディ
type OAuthGoogleRequest struct {
	ProviderUserID string  `json:"providerUserId" binding:"required"`
	Email          string  `json:"email" binding:"required,email"`
	DisplayName    string  `json:"displayName" binding:"required,max=20"`
	AccessToken    string  `json:"accessToken" binding:"required"`
	RefreshToken   *string `json:"refreshToken"`
	ExpiresAt      *int64  `json:"expiresAt"`
}
