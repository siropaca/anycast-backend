package di

import (
	"context"
	"log"

	"gorm.io/gorm"

	"github.com/siropaca/anycast-backend/internal/config"
	"github.com/siropaca/anycast-backend/internal/handler"
	"github.com/siropaca/anycast-backend/internal/infrastructure/cloudtasks"
	"github.com/siropaca/anycast-backend/internal/infrastructure/llm"
	"github.com/siropaca/anycast-backend/internal/infrastructure/storage"
	"github.com/siropaca/anycast-backend/internal/infrastructure/tts"
	"github.com/siropaca/anycast-backend/internal/infrastructure/websocket"
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
	AudioHandler      *handler.AudioHandler
	BgmHandler        *handler.BgmHandler
	AudioJobHandler   *handler.AudioJobHandler
	WorkerHandler     *handler.WorkerHandler
	WebSocketHandler  *handler.WebSocketHandler
	TokenManager      jwt.TokenManager
	UserRepository    repository.UserRepository
	WebSocketHub      *websocket.Hub
}

// 依存関係を構築して Container を返す
func NewContainer(ctx context.Context, db *gorm.DB, cfg *config.Config) *Container {
	// Pkg
	passwordHasher := crypto.NewPasswordHasher()
	tokenManager := jwt.NewTokenManager(cfg.AuthSecret)

	// Infrastructure
	llmClient := llm.NewOpenAIClient(cfg.OpenAIAPIKey)

	// Storage クライアント（GCS）
	storageClient, err := storage.NewGCSClient(ctx, cfg.GoogleCloudStorageBucketName, cfg.GoogleCloudCredentialsJSON)
	if err != nil {
		log.Fatalf("failed to create storage client: %v", err)
	}

	// TTS クライアント（音声生成用）
	ttsClient, err := tts.NewGoogleTTSClient(ctx, cfg.GoogleCloudCredentialsJSON)
	if err != nil {
		log.Fatalf("failed to create TTS client: %v", err)
	}

	// Cloud Tasks クライアント
	var tasksClient cloudtasks.Client
	if cfg.GoogleCloudProjectID != "" && cfg.GoogleCloudTasksWorkerURL != "" {
		tasksClient, err = cloudtasks.NewClient(ctx, cloudtasks.Config{
			ProjectID:           cfg.GoogleCloudProjectID,
			Location:            cfg.GoogleCloudTasksLocation,
			QueueName:           cfg.GoogleCloudTasksQueueName,
			ServiceAccountEmail: cfg.GoogleCloudTasksServiceAccountEmail,
			WorkerEndpointURL:   cfg.GoogleCloudTasksWorkerURL,
			CredentialsJSON:     cfg.GoogleCloudCredentialsJSON,
		})
		if err != nil {
			log.Fatalf("failed to create cloud tasks client: %v", err)
		}
	}

	// WebSocket Hub
	wsHub := websocket.NewHub()
	go wsHub.Run()

	// FFmpeg サービス
	ffmpegService := service.NewFFmpegService()

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
	systemBgmRepo := repository.NewSystemBgmRepository(db)
	audioJobRepo := repository.NewAudioJobRepository(db)

	// Service 層
	voiceService := service.NewVoiceService(voiceRepo)
	authService := service.NewAuthService(userRepo, credentialRepo, oauthAccountRepo, imageRepo, passwordHasher, storageClient)
	channelService := service.NewChannelService(db, channelRepo, characterRepo, categoryRepo, imageRepo, voiceRepo, episodeRepo, bgmRepo, systemBgmRepo, storageClient)
	characterService := service.NewCharacterService(characterRepo, voiceRepo, imageRepo, storageClient)
	categoryService := service.NewCategoryService(categoryRepo)
	episodeService := service.NewEpisodeService(episodeRepo, channelRepo, scriptLineRepo, audioRepo, imageRepo, bgmRepo, systemBgmRepo, storageClient, ttsClient)
	scriptLineService := service.NewScriptLineService(db, scriptLineRepo, episodeRepo, channelRepo)
	scriptService := service.NewScriptService(db, userRepo, channelRepo, episodeRepo, scriptLineRepo, llmClient, storageClient)
	cleanupService := service.NewCleanupService(audioRepo, imageRepo, storageClient)
	imageService := service.NewImageService(imageRepo, storageClient)
	audioService := service.NewAudioService(audioRepo, storageClient)
	bgmService := service.NewBgmService(bgmRepo, systemBgmRepo, audioRepo, storageClient)
	audioJobService := service.NewAudioJobService(
		audioJobRepo,
		episodeRepo,
		channelRepo,
		scriptLineRepo,
		audioRepo,
		bgmRepo,
		systemBgmRepo,
		storageClient,
		ttsClient,
		ffmpegService,
		tasksClient,
		wsHub,
	)

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
	audioHandler := handler.NewAudioHandler(audioService)
	bgmHandler := handler.NewBgmHandler(bgmService)
	audioJobHandler := handler.NewAudioJobHandler(audioJobService)
	workerHandler := handler.NewWorkerHandler(audioJobService)
	webSocketHandler := handler.NewWebSocketHandler(wsHub, tokenManager)

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
		AudioHandler:      audioHandler,
		BgmHandler:        bgmHandler,
		AudioJobHandler:   audioJobHandler,
		WorkerHandler:     workerHandler,
		WebSocketHandler:  webSocketHandler,
		TokenManager:      tokenManager,
		UserRepository:    userRepo,
		WebSocketHub:      wsHub,
	}
}
