package di

import (
	"context"
	"os"

	"gorm.io/gorm"

	"github.com/siropaca/anycast-backend/internal/config"
	"github.com/siropaca/anycast-backend/internal/handler"
	"github.com/siropaca/anycast-backend/internal/infrastructure/cloudtasks"
	"github.com/siropaca/anycast-backend/internal/infrastructure/imagegen"
	"github.com/siropaca/anycast-backend/internal/infrastructure/llm"
	"github.com/siropaca/anycast-backend/internal/infrastructure/slack"
	"github.com/siropaca/anycast-backend/internal/infrastructure/storage"
	"github.com/siropaca/anycast-backend/internal/infrastructure/tts"
	"github.com/siropaca/anycast-backend/internal/infrastructure/websocket"
	"github.com/siropaca/anycast-backend/internal/pkg/crypto"
	"github.com/siropaca/anycast-backend/internal/pkg/jwt"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
	"github.com/siropaca/anycast-backend/internal/pkg/tracer"
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
	SearchHandler          *handler.SearchHandler
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
	log := logger.Default()

	llmRegistry := llm.NewRegistry()

	// OpenAI（API キーがあれば登録）
	if cfg.OpenAIAPIKey != "" {
		openaiClient, err := llm.NewClient(llm.ClientConfig{
			Provider:     llm.ProviderOpenAI,
			OpenAIAPIKey: cfg.OpenAIAPIKey,
		})
		if err != nil {
			log.Error("failed to create OpenAI client", "error", err)
			os.Exit(1)
		}
		llmRegistry.Register(llm.ProviderOpenAI, openaiClient)
		log.Info("LLM provider registered", "provider", "openai")
	}

	// Claude（API キーがあれば登録）
	if cfg.ClaudeAPIKey != "" {
		claudeClient, err := llm.NewClient(llm.ClientConfig{
			Provider:     llm.ProviderClaude,
			ClaudeAPIKey: cfg.ClaudeAPIKey,
		})
		if err != nil {
			log.Error("failed to create Claude client", "error", err)
			os.Exit(1)
		}
		llmRegistry.Register(llm.ProviderClaude, claudeClient)
		log.Info("LLM provider registered", "provider", "claude")
	}

	// Gemini（プロジェクト ID があれば登録）
	if cfg.GoogleCloudProjectID != "" {
		geminiClient, err := llm.NewClient(llm.ClientConfig{
			Provider:          llm.ProviderGemini,
			GeminiProjectID:   cfg.GoogleCloudProjectID,
			GeminiLocation:    cfg.GeminiLLMLocation,
			GeminiCredentials: cfg.GoogleCloudCredentialsJSON,
		})
		if err != nil {
			log.Error("failed to create Gemini client", "error", err)
			os.Exit(1)
		}
		llmRegistry.Register(llm.ProviderGemini, geminiClient)
		log.Info("LLM provider registered", "provider", "gemini")
	}

	// Phase 設定で使用するプロバイダが登録されているかバリデーション
	for _, pc := range service.PhaseConfigs() {
		if !llmRegistry.Has(pc.Provider) {
			log.Error("required LLM provider is not configured", "provider", pc.Provider)
			os.Exit(1)
		}
	}

	// Storage クライアント（GCS）
	storageClient, err := storage.NewGCSClient(ctx, cfg.GoogleCloudStorageBucketName, cfg.GoogleCloudCredentialsJSON)
	if err != nil {
		log.Error("failed to create storage client", "error", err)
		os.Exit(1)
	}

	// TTS クライアント（レジストリパターン）
	ttsRegistry := tts.NewRegistry()

	// Gemini（プロジェクト ID があれば登録）
	if cfg.GoogleCloudProjectID != "" {
		geminiTTSClient, err := tts.NewGeminiTTSClient(ctx, cfg.GoogleCloudProjectID, cfg.GoogleCloudTTSLocation, cfg.GoogleCloudCredentialsJSON)
		if err != nil {
			log.Error("failed to create Gemini TTS client", "error", err)
			os.Exit(1)
		}
		ttsRegistry.Register(tts.ProviderGoogle, geminiTTSClient)
		log.Info("TTS provider registered", "provider", "gemini")
	}

	// ElevenLabs（API キーがあれば登録）
	if cfg.ElevenLabsAPIKey != "" {
		elevenLabsTTSClient := tts.NewElevenLabsTTSClient(cfg.ElevenLabsAPIKey)
		ttsRegistry.Register(tts.ProviderElevenLabs, elevenLabsTTSClient)
		log.Info("TTS provider registered", "provider", "elevenlabs")
	}

	log.Info("TTS providers registered", "providers", ttsRegistry.Providers())

	// 画像生成クライアント（レジストリパターン）
	imagegenRegistry := imagegen.NewRegistry()

	// Gemini（プロジェクト ID があれば登録）
	if cfg.GoogleCloudProjectID != "" {
		geminiImagegenClient, err := imagegen.NewGeminiClient(ctx, cfg.GoogleCloudProjectID, cfg.GeminiImageGenLocation, cfg.GoogleCloudCredentialsJSON)
		if err != nil {
			log.Error("failed to create Gemini image gen client", "error", err)
			os.Exit(1)
		}
		imagegenRegistry.Register(imagegen.ProviderGemini, geminiImagegenClient)
		log.Info("ImageGen provider registered", "provider", "gemini")
	}

	// OpenAI（API キーがあれば登録）
	if cfg.OpenAIAPIKey != "" {
		openaiImagegenClient := imagegen.NewOpenAIClient(cfg.OpenAIAPIKey, cfg.OpenAIImageGenModel)
		imagegenRegistry.Register(imagegen.ProviderOpenAI, openaiImagegenClient)
		log.Info("ImageGen provider registered", "provider", "openai")
	}

	// 指定プロバイダのクライアントを取得
	imagegenProvider := imagegen.Provider(cfg.ImageGenProvider)
	imagegenClient, err := imagegenRegistry.Get(imagegenProvider)
	if err != nil {
		log.Error("image gen provider is not configured", "provider", cfg.ImageGenProvider)
		os.Exit(1)
	}
	log.Info("ImageGen provider selected", "provider", cfg.ImageGenProvider)

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
	slackClient := slack.NewClient(cfg.SlackFeedbackWebhookURL, cfg.SlackContactWebhookURL, cfg.SlackAlertWebhookURL, cfg.SlackRegistrationWebhookURL)

	// FFmpeg サービス
	ffmpegService := service.NewFFmpegService()

	// Repository 層
	voiceRepo := repository.NewVoiceRepository(db)
	favVoiceRepo := repository.NewFavoriteVoiceRepository(db)
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
	voiceService := service.NewVoiceService(voiceRepo, favVoiceRepo)
	authService := service.NewAuthService(userRepo, credentialRepo, oauthAccountRepo, refreshTokenRepo, imageRepo, playlistRepo, audioJobRepo, scriptJobRepo, passwordHasher, storageClient, slackClient)
	channelService := service.NewChannelService(db, channelRepo, characterRepo, categoryRepo, imageRepo, voiceRepo, episodeRepo, scriptLineRepo, bgmRepo, systemBgmRepo, playbackHistoryRepo, storageClient)
	characterService := service.NewCharacterService(characterRepo, voiceRepo, imageRepo, storageClient)
	categoryService := service.NewCategoryService(categoryRepo, storageClient)
	episodeService := service.NewEpisodeService(episodeRepo, channelRepo, scriptLineRepo, audioRepo, imageRepo, bgmRepo, systemBgmRepo, playbackHistoryRepo, playlistRepo, storageClient, ttsRegistry)
	scriptLineService := service.NewScriptLineService(db, scriptLineRepo, episodeRepo, channelRepo)
	scriptService := service.NewScriptService(db, channelRepo, episodeRepo, scriptLineRepo, storageClient)
	cleanupService := service.NewCleanupService(audioRepo, imageRepo, storageClient)
	imageService := service.NewImageService(imageRepo, storageClient, imagegenClient)
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
		ttsRegistry,
		ffmpegService,
		tasksClient,
		wsHub,
		slackClient,
	)
	scriptJobService := service.NewScriptJobService(
		db,
		scriptJobRepo,
		userRepo,
		channelRepo,
		episodeRepo,
		scriptLineRepo,
		llmRegistry,
		tasksClient,
		wsHub,
		tracer.Mode(cfg.TraceMode),
		slackClient,
	)
	feedbackService := service.NewFeedbackService(feedbackRepo, imageRepo, userRepo, storageClient, slackClient)
	contactService := service.NewContactService(contactRepo, slackClient)
	playlistService := service.NewPlaylistService(db, playlistRepo, episodeRepo, storageClient)
	playbackHistoryService := service.NewPlaybackHistoryService(playbackHistoryRepo, episodeRepo, storageClient)
	followService := service.NewFollowService(followRepo, userRepo, storageClient)
	reactionService := service.NewReactionService(reactionRepo, storageClient)
	recommendationService := service.NewRecommendationService(recommendationRepo, categoryRepo, storageClient)
	searchService := service.NewSearchService(channelRepo, episodeRepo, userRepo, storageClient)
	userService := service.NewUserService(userRepo, channelRepo, episodeRepo, followRepo, storageClient)
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
	searchHandler := handler.NewSearchHandler(searchService)
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
		SearchHandler:          searchHandler,
		UserHandler:            userHandler,
		TokenManager:           tokenManager,
		UserRepository:         userRepo,
		WebSocketHub:           wsHub,
	}
}
