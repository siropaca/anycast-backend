package di

import (
	"context"
	"log"

	"gorm.io/gorm"

	"github.com/siropaca/anycast-backend/internal/config"
	"github.com/siropaca/anycast-backend/internal/handler"
	"github.com/siropaca/anycast-backend/internal/infrastructure/llm"
	"github.com/siropaca/anycast-backend/internal/infrastructure/storage"
	"github.com/siropaca/anycast-backend/internal/infrastructure/tts"
	"github.com/siropaca/anycast-backend/internal/pkg/crypto"
	"github.com/siropaca/anycast-backend/internal/pkg/jwt"
	"github.com/siropaca/anycast-backend/internal/repository"
	"github.com/siropaca/anycast-backend/internal/service"
)

// DI コンテナ
type Container struct {
	VoiceHandler      *handler.VoiceHandler
	AuthHandler       *handler.AuthHandler
	ChannelHandler    *handler.ChannelHandler
	CharacterHandler  *handler.CharacterHandler
	CategoryHandler   *handler.CategoryHandler
	EpisodeHandler    *handler.EpisodeHandler
	ScriptLineHandler *handler.ScriptLineHandler
	ScriptHandler     *handler.ScriptHandler
	CleanupHandler    *handler.CleanupHandler
	ImageHandler      *handler.ImageHandler
	BgmHandler        *handler.BgmHandler
	TokenManager      jwt.TokenManager
	UserRepository    repository.UserRepository
}

// 依存関係を構築して Container を返す
func NewContainer(ctx context.Context, db *gorm.DB, cfg *config.Config) *Container {
	// Pkg
	passwordHasher := crypto.NewPasswordHasher()
	tokenManager := jwt.NewTokenManager(cfg.AuthSecret)

	// Infrastructure
	llmClient := llm.NewOpenAIClient(cfg.OpenAIAPIKey)

	// Storage クライアント（GCS）
	storageClient, err := storage.NewGCSClient(ctx, cfg.GCSBucketName, cfg.GoogleCredentialsJSON)
	if err != nil {
		log.Fatalf("failed to create storage client: %v", err)
	}

	// TTS クライアント（音声生成用）
	ttsClient, err := tts.NewGoogleTTSClient(ctx, cfg.GoogleCredentialsJSON)
	if err != nil {
		log.Fatalf("failed to create TTS client: %v", err)
	}

	// Repository 層
	voiceRepo := repository.NewVoiceRepository(db)
	userRepo := repository.NewUserRepository(db)
	credentialRepo := repository.NewCredentialRepository(db)
	oauthAccountRepo := repository.NewOAuthAccountRepository(db)
	imageRepo := repository.NewImageRepository(db)
	channelRepo := repository.NewChannelRepository(db)
	characterRepo := repository.NewCharacterRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	episodeRepo := repository.NewEpisodeRepository(db)
	scriptLineRepo := repository.NewScriptLineRepository(db)
	audioRepo := repository.NewAudioRepository(db)
	bgmRepo := repository.NewBgmRepository(db)
	defaultBgmRepo := repository.NewDefaultBgmRepository(db)

	// Service 層
	voiceService := service.NewVoiceService(voiceRepo)
	authService := service.NewAuthService(userRepo, credentialRepo, oauthAccountRepo, imageRepo, passwordHasher, storageClient)
	channelService := service.NewChannelService(db, channelRepo, characterRepo, categoryRepo, imageRepo, voiceRepo, episodeRepo, storageClient)
	characterService := service.NewCharacterService(characterRepo, voiceRepo, imageRepo, storageClient)
	categoryService := service.NewCategoryService(categoryRepo)
	episodeService := service.NewEpisodeService(episodeRepo, channelRepo, scriptLineRepo, audioRepo, imageRepo, storageClient, ttsClient)
	scriptLineService := service.NewScriptLineService(db, scriptLineRepo, episodeRepo, channelRepo)
	scriptService := service.NewScriptService(db, userRepo, channelRepo, episodeRepo, scriptLineRepo, llmClient, storageClient)
	cleanupService := service.NewCleanupService(audioRepo, imageRepo, storageClient)
	imageService := service.NewImageService(imageRepo, storageClient)
	bgmService := service.NewBgmService(bgmRepo, defaultBgmRepo, audioRepo, storageClient)

	// Handler 層
	voiceHandler := handler.NewVoiceHandler(voiceService)
	authHandler := handler.NewAuthHandler(authService, tokenManager)
	channelHandler := handler.NewChannelHandler(channelService)
	characterHandler := handler.NewCharacterHandler(characterService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	episodeHandler := handler.NewEpisodeHandler(episodeService)
	scriptLineHandler := handler.NewScriptLineHandler(scriptLineService)
	scriptHandler := handler.NewScriptHandler(scriptService)
	cleanupHandler := handler.NewCleanupHandler(cleanupService, storageClient)
	imageHandler := handler.NewImageHandler(imageService)
	bgmHandler := handler.NewBgmHandler(bgmService)

	return &Container{
		VoiceHandler:      voiceHandler,
		AuthHandler:       authHandler,
		ChannelHandler:    channelHandler,
		CharacterHandler:  characterHandler,
		CategoryHandler:   categoryHandler,
		EpisodeHandler:    episodeHandler,
		ScriptLineHandler: scriptLineHandler,
		ScriptHandler:     scriptHandler,
		CleanupHandler:    cleanupHandler,
		ImageHandler:      imageHandler,
		BgmHandler:        bgmHandler,
		TokenManager:      tokenManager,
		UserRepository:    userRepo,
	}
}
