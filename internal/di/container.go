package di

import (
	"context"
	"os"

	"gorm.io/gorm"

	"github.com/siropaca/anycast-backend/internal/config"
	"github.com/siropaca/anycast-backend/internal/handler"
	"github.com/siropaca/anycast-backend/internal/infrastructure/cloudtasks"
	"github.com/siropaca/anycast-backend/internal/infrastructure/llm"
	"github.com/siropaca/anycast-backend/internal/infrastructure/slack"
	"github.com/siropaca/anycast-backend/internal/infrastructure/storage"
	"github.com/siropaca/anycast-backend/internal/infrastructure/tts"
	"github.com/siropaca/anycast-backend/internal/infrastructure/websocket"
	"github.com/siropaca/anycast-backend/internal/pkg/crypto"
	"github.com/siropaca/anycast-backend/internal/pkg/jwt"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
	"github.com/siropaca/anycast-backend/internal/repository"
	"github.com/siropaca/anycast-backend/internal/service"
)

// DI コンテナ
type Container struct {
	VoiceHandler           *handler.VoiceHandler
	AuthHandler            *handler.AuthHandler
	ChannelHandler         *handler.ChannelHandler
	CharacterHandler       *handler.CharacterHandler
	CategoryHandler        *handler.CategoryHandler
	EpisodeHandler         *handler.EpisodeHandler
	ScriptLineHandler      *handler.ScriptLineHandler
	ScriptHandler          *handler.ScriptHandler
	ScriptJobHandler       *handler.ScriptJobHandler
	CleanupHandler         *handler.CleanupHandler
	ImageHandler           *handler.ImageHandler
	AudioHandler           *handler.AudioHandler
	BgmHandler             *handler.BgmHandler
	AudioJobHandler        *handler.AudioJobHandler
	WorkerHandler          *handler.WorkerHandler
	WebSocketHandler       *handler.WebSocketHandler
	FeedbackHandler        *handler.FeedbackHandler
	ContactHandler         *handler.ContactHandler
	PlaylistHandler        *handler.PlaylistHandler
	PlaybackHistoryHandler *handler.PlaybackHistoryHandler
	FollowHandler          *handler.FollowHandler
	ReactionHandler        *handler.ReactionHandler
	RecommendationHandler  *handler.RecommendationHandler
	UserHandler            *handler.UserHandler
	TokenManager           jwt.TokenManager
	UserRepository         repository.UserRepository
	WebSocketHub           *websocket.Hub
}

// 依存関係を構築して Container を返す
func NewContainer(ctx context.Context, db *gorm.DB, cfg *config.Config) *Container {
	// Pkg
	passwordHasher := crypto.NewPasswordHasher()
	tokenManager := jwt.NewTokenManager(cfg.AuthSecret)

	// Infrastructure
	llmClient := llm.NewOpenAIClient(cfg.OpenAIAPIKey)

	log := logger.Default()

	// Storage クライアント（GCS）
	storageClient, err := storage.NewGCSClient(ctx, cfg.GoogleCloudStorageBucketName, cfg.GoogleCloudCredentialsJSON)
	if err != nil {
		log.Error("failed to create storage client", "error", err)
		os.Exit(1)
	}

	// TTS クライアント（Gemini TTS - 32k token の長い台本をサポート）
	log.Debug("TTS config", "project_id", cfg.GoogleCloudProjectID, "location", cfg.GoogleCloudTTSLocation)
	if cfg.GoogleCloudProjectID == "" {
		log.Error("GOOGLE_CLOUD_PROJECT_ID is required for Gemini TTS")
		os.Exit(1)
	}
	ttsClient, err := tts.NewGeminiTTSClient(ctx, cfg.GoogleCloudProjectID, cfg.GoogleCloudTTSLocation, cfg.GoogleCloudCredentialsJSON)
	if err != nil {
		log.Error("failed to create Gemini TTS client", "error", err)
		os.Exit(1)
	}
	log.Debug("using Gemini TTS client (32k token support)")

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
			log.Error("failed to create cloud tasks client", "error", err)
			os.Exit(1)
		}
	}

	// WebSocket Hub
	wsHub := websocket.NewHub()
	go wsHub.Run()

	// Slack クライアント
	slackClient := slack.NewClient(cfg.SlackWebhookURL)

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
	scriptJobRepo := repository.NewScriptJobRepository(db)
	feedbackRepo := repository.NewFeedbackRepository(db)
	contactRepo := repository.NewContactRepository(db)
	playlistRepo := repository.NewPlaylistRepository(db)
	playbackHistoryRepo := repository.NewPlaybackHistoryRepository(db)
	followRepo := repository.NewFollowRepository(db)
	reactionRepo := repository.NewReactionRepository(db)
	recommendationRepo := repository.NewRecommendationRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)

	// Service 層
	voiceService := service.NewVoiceService(voiceRepo)
	authService := service.NewAuthService(userRepo, credentialRepo, oauthAccountRepo, refreshTokenRepo, imageRepo, playlistRepo, passwordHasher, storageClient)
	channelService := service.NewChannelService(db, channelRepo, characterRepo, categoryRepo, imageRepo, voiceRepo, episodeRepo, bgmRepo, systemBgmRepo, storageClient)
	characterService := service.NewCharacterService(characterRepo, voiceRepo, imageRepo, storageClient)
	categoryService := service.NewCategoryService(categoryRepo, storageClient)
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
	scriptJobService := service.NewScriptJobService(
		db,
		scriptJobRepo,
		userRepo,
		channelRepo,
		episodeRepo,
		scriptLineRepo,
		llmClient,
		tasksClient,
		wsHub,
	)
	feedbackService := service.NewFeedbackService(feedbackRepo, imageRepo, userRepo, storageClient, slackClient)
	contactService := service.NewContactService(contactRepo, slackClient)
	playlistService := service.NewPlaylistService(playlistRepo, episodeRepo, storageClient)
	playbackHistoryService := service.NewPlaybackHistoryService(playbackHistoryRepo, episodeRepo, storageClient)
	followService := service.NewFollowService(followRepo, storageClient)
	reactionService := service.NewReactionService(reactionRepo, storageClient)
	recommendationService := service.NewRecommendationService(recommendationRepo, categoryRepo, storageClient)
	userService := service.NewUserService(userRepo, channelRepo, storageClient)

	// Handler 層
	voiceHandler := handler.NewVoiceHandler(voiceService)
	authHandler := handler.NewAuthHandler(authService, tokenManager)
	channelHandler := handler.NewChannelHandler(channelService)
	characterHandler := handler.NewCharacterHandler(characterService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	episodeHandler := handler.NewEpisodeHandler(episodeService)
	scriptLineHandler := handler.NewScriptLineHandler(scriptLineService)
	scriptHandler := handler.NewScriptHandler(scriptService)
	scriptJobHandler := handler.NewScriptJobHandler(scriptJobService)
	cleanupHandler := handler.NewCleanupHandler(cleanupService, storageClient)
	imageHandler := handler.NewImageHandler(imageService)
	audioHandler := handler.NewAudioHandler(audioService)
	bgmHandler := handler.NewBgmHandler(bgmService)
	audioJobHandler := handler.NewAudioJobHandler(audioJobService)
	workerHandler := handler.NewWorkerHandler(audioJobService, scriptJobService)
	webSocketHandler := handler.NewWebSocketHandler(wsHub, tokenManager)
	feedbackHandler := handler.NewFeedbackHandler(feedbackService)
	contactHandler := handler.NewContactHandler(contactService)
	playlistHandler := handler.NewPlaylistHandler(playlistService)
	playbackHistoryHandler := handler.NewPlaybackHistoryHandler(playbackHistoryService)
	followHandler := handler.NewFollowHandler(followService)
	reactionHandler := handler.NewReactionHandler(reactionService)
	recommendationHandler := handler.NewRecommendationHandler(recommendationService)
	userHandler := handler.NewUserHandler(userService)

	return &Container{
		VoiceHandler:           voiceHandler,
		AuthHandler:            authHandler,
		ChannelHandler:         channelHandler,
		CharacterHandler:       characterHandler,
		CategoryHandler:        categoryHandler,
		EpisodeHandler:         episodeHandler,
		ScriptLineHandler:      scriptLineHandler,
		ScriptHandler:          scriptHandler,
		ScriptJobHandler:       scriptJobHandler,
		CleanupHandler:         cleanupHandler,
		ImageHandler:           imageHandler,
		AudioHandler:           audioHandler,
		BgmHandler:             bgmHandler,
		AudioJobHandler:        audioJobHandler,
		WorkerHandler:          workerHandler,
		WebSocketHandler:       webSocketHandler,
		FeedbackHandler:        feedbackHandler,
		ContactHandler:         contactHandler,
		PlaylistHandler:        playlistHandler,
		PlaybackHistoryHandler: playbackHistoryHandler,
		FollowHandler:          followHandler,
		ReactionHandler:        reactionHandler,
		RecommendationHandler:  recommendationHandler,
		UserHandler:            userHandler,
		TokenManager:           tokenManager,
		UserRepository:         userRepo,
		WebSocketHub:           wsHub,
	}
}
