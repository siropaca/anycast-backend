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
	VoiceHandler    *handler.VoiceHandler
	AuthHandler     *handler.AuthHandler
	ChannelHandler  *handler.ChannelHandler
	CategoryHandler *handler.CategoryHandler
	EpisodeHandler  *handler.EpisodeHandler
	TokenManager    jwt.TokenManager
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
	imageRepo := repository.NewImageRepository(db)
	channelRepo := repository.NewChannelRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	episodeRepo := repository.NewEpisodeRepository(db)

	// Service 層
	voiceService := service.NewVoiceService(voiceRepo)
	authService := service.NewAuthService(userRepo, credentialRepo, oauthAccountRepo, imageRepo, passwordHasher)
	channelService := service.NewChannelService(channelRepo, categoryRepo, imageRepo, voiceRepo)
	categoryService := service.NewCategoryService(categoryRepo)
	episodeService := service.NewEpisodeService(episodeRepo, channelRepo)

	// Handler 層
	voiceHandler := handler.NewVoiceHandler(voiceService)
	authHandler := handler.NewAuthHandler(authService, tokenManager)
	channelHandler := handler.NewChannelHandler(channelService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	episodeHandler := handler.NewEpisodeHandler(episodeService)

	return &Container{
		VoiceHandler:    voiceHandler,
		AuthHandler:     authHandler,
		ChannelHandler:  channelHandler,
		CategoryHandler: categoryHandler,
		EpisodeHandler:  episodeHandler,
		TokenManager:    tokenManager,
	}
}
