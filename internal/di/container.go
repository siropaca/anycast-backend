package di

import (
	"gorm.io/gorm"

	"github.com/siropaca/anycast-backend/internal/config"
	"github.com/siropaca/anycast-backend/internal/handler"
	"github.com/siropaca/anycast-backend/internal/pkg/crypto"
	"github.com/siropaca/anycast-backend/internal/pkg/jwt"
	"github.com/siropaca/anycast-backend/internal/repository"
	"github.com/siropaca/anycast-backend/internal/service"
)

// DI コンテナ
type Container struct {
	VoiceHandler *handler.VoiceHandler
	AuthHandler  *handler.AuthHandler
	TokenManager jwt.TokenManager
}

// 依存関係を構築して Container を返す
func NewContainer(db *gorm.DB, cfg *config.Config) *Container {
	// Pkg
	passwordHasher := crypto.NewPasswordHasher()
	tokenManager := jwt.NewTokenManager(cfg.AuthSecret)

	// Repository 層
	voiceRepo := repository.NewVoiceRepository(db)
	userRepo := repository.NewUserRepository(db)
	credentialRepo := repository.NewCredentialRepository(db)
	oauthAccountRepo := repository.NewOAuthAccountRepository(db)

	// Service 層
	voiceService := service.NewVoiceService(voiceRepo)
	authService := service.NewAuthService(userRepo, credentialRepo, oauthAccountRepo, passwordHasher)

	// Handler 層
	voiceHandler := handler.NewVoiceHandler(voiceService)
	authHandler := handler.NewAuthHandler(authService)

	return &Container{
		VoiceHandler: voiceHandler,
		AuthHandler:  authHandler,
		TokenManager: tokenManager,
	}
}
