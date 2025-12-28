package service

import (
	"context"

	"github.com/siropaca/anycast-backend/internal/dto/request"
	"github.com/siropaca/anycast-backend/internal/dto/response"
	"github.com/siropaca/anycast-backend/internal/model"
	"github.com/siropaca/anycast-backend/internal/repository"
)

// ボイス関連のビジネスロジックインターフェース
type VoiceService interface {
	ListVoices(ctx context.Context, filter repository.VoiceFilter) ([]model.Voice, error)
	GetVoice(ctx context.Context, id string) (*model.Voice, error)
}

// 認証結果
type AuthResult struct {
	User      response.UserResponse
	IsCreated bool // 新規作成されたかどうか（OAuth 用）
}

// 認証関連のビジネスロジックインターフェース
type AuthService interface {
	Register(ctx context.Context, req request.RegisterRequest) (*response.UserResponse, error)
	Login(ctx context.Context, req request.LoginRequest) (*response.UserResponse, error)
	OAuthGoogle(ctx context.Context, req request.OAuthGoogleRequest) (*AuthResult, error)
}
