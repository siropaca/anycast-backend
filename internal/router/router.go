package router

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/siropaca/anycast-backend/internal/config"
	"github.com/siropaca/anycast-backend/internal/di"
	"github.com/siropaca/anycast-backend/internal/middleware"
	_ "github.com/siropaca/anycast-backend/swagger"
)

// Setup はルーターを設定して返す
func Setup(container *di.Container, cfg *config.Config) *gin.Engine {
	if cfg.AppEnv == config.EnvProduction {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// ミドルウェア
	r.Use(middleware.Logger())
	r.Use(middleware.ErrorHandler())
	r.Use(gin.Recovery())

	// CORS 設定
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSAllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Swagger（本番環境では無効）
	if cfg.AppEnv != config.EnvProduction {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// ルートエンドポイント
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to Anycast API",
		})
	})

	// ヘルスチェック
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// API v1
	api := r.Group("/api/v1")

	// Auth（認証不要）
	auth := api.Group("/auth")
	auth.POST("/register", container.AuthHandler.Register)
	auth.POST("/login", container.AuthHandler.Login)
	auth.POST("/oauth/google", container.AuthHandler.OAuthGoogle)
	auth.POST("/refresh", container.AuthHandler.RefreshToken)

	// 認証必須のエンドポイント
	authenticated := api.Group("")
	authenticated.Use(middleware.Auth(container.TokenManager))

	// Auth（認証必須）
	authenticated.POST("/auth/logout", container.AuthHandler.Logout)

	// Me（自分のリソース）
	authenticated.GET("/me", container.AuthHandler.GetMe)
	authenticated.PATCH("/me/prompt", container.AuthHandler.UpdatePrompt)
	authenticated.GET("/me/channels", container.ChannelHandler.ListMyChannels)
	authenticated.GET("/me/channels/:channelId", container.ChannelHandler.GetMyChannel)
	authenticated.GET("/me/channels/:channelId/episodes", container.EpisodeHandler.ListMyChannelEpisodes)
	authenticated.GET("/me/channels/:channelId/episodes/:episodeId", container.EpisodeHandler.GetMyChannelEpisode)
	authenticated.GET("/me/characters", container.CharacterHandler.ListMyCharacters)
	authenticated.GET("/me/characters/:characterId", container.CharacterHandler.GetMyCharacter)
	authenticated.POST("/me/characters", container.CharacterHandler.CreateCharacter)
	authenticated.PATCH("/me/characters/:characterId", container.CharacterHandler.UpdateCharacter)
	authenticated.DELETE("/me/characters/:characterId", container.CharacterHandler.DeleteCharacter)
	authenticated.GET("/me/bgms", container.BgmHandler.ListMyBgms)
	authenticated.POST("/me/bgms", container.BgmHandler.CreateBgm)
	authenticated.GET("/me/bgms/:bgmId", container.BgmHandler.GetMyBgm)
	authenticated.PATCH("/me/bgms/:bgmId", container.BgmHandler.UpdateMyBgm)
	authenticated.DELETE("/me/bgms/:bgmId", container.BgmHandler.DeleteMyBgm)
	authenticated.GET("/me/audio-jobs", container.AudioJobHandler.ListMyAudioJobs)
	authenticated.GET("/me/script-jobs", container.ScriptJobHandler.ListMyScriptJobs)

	// Playlists
	authenticated.GET("/me/playlists", container.PlaylistHandler.ListPlaylists)
	authenticated.GET("/me/playlists/:playlistId", container.PlaylistHandler.GetPlaylist)
	authenticated.POST("/me/playlists", container.PlaylistHandler.CreatePlaylist)
	authenticated.PATCH("/me/playlists/:playlistId", container.PlaylistHandler.UpdatePlaylist)
	authenticated.DELETE("/me/playlists/:playlistId", container.PlaylistHandler.DeletePlaylist)
	authenticated.POST("/me/playlists/:playlistId/items", container.PlaylistHandler.AddItem)
	authenticated.DELETE("/me/playlists/:playlistId/items/:itemId", container.PlaylistHandler.RemoveItem)
	authenticated.POST("/me/playlists/:playlistId/items/reorder", container.PlaylistHandler.ReorderItems)
	authenticated.GET("/me/default-playlist", container.PlaylistHandler.GetDefaultPlaylist)

	// Playback History
	authenticated.GET("/me/playback-history", container.PlaybackHistoryHandler.ListPlaybackHistory)

	// Follows
	authenticated.GET("/me/follows", container.FollowHandler.ListFollows)
	authenticated.GET("/users/:username/follow", container.FollowHandler.GetFollowStatus)
	authenticated.POST("/users/:username/follow", container.FollowHandler.CreateFollow)
	authenticated.DELETE("/users/:username/follow", container.FollowHandler.DeleteFollow)

	// Users
	authenticated.GET("/users/:username", container.UserHandler.GetUser)

	// Likes
	authenticated.GET("/me/likes", container.ReactionHandler.ListLikes)

	// Channels
	authenticated.POST("/channels", container.ChannelHandler.CreateChannel)
	authenticated.PATCH("/channels/:channelId", container.ChannelHandler.UpdateChannel)
	authenticated.DELETE("/channels/:channelId", container.ChannelHandler.DeleteChannel)
	authenticated.POST("/channels/:channelId/publish", container.ChannelHandler.PublishChannel)
	authenticated.POST("/channels/:channelId/unpublish", container.ChannelHandler.UnpublishChannel)
	authenticated.DELETE("/channels/:channelId/default-bgm", container.ChannelHandler.DeleteDefaultBgm)

	// Episodes
	authenticated.POST("/channels/:channelId/episodes", container.EpisodeHandler.CreateEpisode)
	authenticated.PATCH("/channels/:channelId/episodes/:episodeId", container.EpisodeHandler.UpdateEpisode)
	authenticated.DELETE("/channels/:channelId/episodes/:episodeId", container.EpisodeHandler.DeleteEpisode)
	authenticated.POST("/channels/:channelId/episodes/:episodeId/publish", container.EpisodeHandler.PublishEpisode)
	authenticated.POST("/channels/:channelId/episodes/:episodeId/unpublish", container.EpisodeHandler.UnpublishEpisode)
	authenticated.PUT("/channels/:channelId/episodes/:episodeId/bgm", container.EpisodeHandler.SetEpisodeBgm)
	authenticated.DELETE("/channels/:channelId/episodes/:episodeId/bgm", container.EpisodeHandler.DeleteEpisodeBgm)
	authenticated.POST("/channels/:channelId/episodes/:episodeId/audio/generate-async", container.AudioJobHandler.GenerateAudioAsync)
	authenticated.POST("/episodes/:episodeId/play", container.EpisodeHandler.IncrementPlayCount)
	authenticated.POST("/episodes/:episodeId/default-playlist", container.PlaylistHandler.AddToDefaultPlaylist)
	authenticated.DELETE("/episodes/:episodeId/default-playlist", container.PlaylistHandler.RemoveFromDefaultPlaylist)
	authenticated.PUT("/episodes/:episodeId/playback", container.PlaybackHistoryHandler.UpdatePlayback)
	authenticated.DELETE("/episodes/:episodeId/playback", container.PlaybackHistoryHandler.DeletePlayback)
	authenticated.POST("/episodes/:episodeId/reactions", container.ReactionHandler.CreateOrUpdateReaction)
	authenticated.DELETE("/episodes/:episodeId/reactions", container.ReactionHandler.DeleteReaction)

	// Audio Jobs
	authenticated.GET("/audio-jobs/:jobId", container.AudioJobHandler.GetAudioJob)
	authenticated.POST("/audio-jobs/:jobId/cancel", container.AudioJobHandler.CancelAudioJob)

	// Script Jobs
	authenticated.GET("/script-jobs/:jobId", container.ScriptJobHandler.GetScriptJob)
	authenticated.POST("/script-jobs/:jobId/cancel", container.ScriptJobHandler.CancelScriptJob)

	// Script Lines
	authenticated.GET("/channels/:channelId/episodes/:episodeId/script/lines", container.ScriptLineHandler.ListScriptLines)
	authenticated.POST("/channels/:channelId/episodes/:episodeId/script/lines", container.ScriptLineHandler.CreateScriptLine)
	authenticated.PATCH("/channels/:channelId/episodes/:episodeId/script/lines/:lineId", container.ScriptLineHandler.UpdateScriptLine)
	authenticated.DELETE("/channels/:channelId/episodes/:episodeId/script/lines/:lineId", container.ScriptLineHandler.DeleteScriptLine)
	authenticated.POST("/channels/:channelId/episodes/:episodeId/script/reorder", container.ScriptLineHandler.ReorderScriptLines)

	// Script（台本）
	authenticated.POST("/channels/:channelId/episodes/:episodeId/script/generate-async", container.ScriptJobHandler.GenerateScriptAsync)
	authenticated.POST("/channels/:channelId/episodes/:episodeId/script/import", container.ScriptHandler.ImportScript)
	authenticated.GET("/channels/:channelId/episodes/:episodeId/script/export", container.ScriptHandler.ExportScript)

	// Categories
	authenticated.GET("/categories/:slug", container.CategoryHandler.GetCategoryBySlug)

	// Voices
	authenticated.GET("/voices", container.VoiceHandler.ListVoices)
	authenticated.GET("/voices/:voiceId", container.VoiceHandler.GetVoice)

	// Images
	authenticated.POST("/images", container.ImageHandler.UploadImage)

	// Audios
	authenticated.POST("/audios", container.AudioHandler.UploadAudio)

	// Feedbacks
	authenticated.POST("/feedbacks", container.FeedbackHandler.CreateFeedback)

	// 任意認証（ログイン時はオーナー判定、未ログインは公開コンテンツのみ）
	optionalAuth := api.Group("")
	optionalAuth.Use(middleware.OptionalAuth(container.TokenManager))
	optionalAuth.GET("/channels/:channelId", container.ChannelHandler.GetChannel)
	optionalAuth.GET("/channels/:channelId/episodes", container.EpisodeHandler.ListChannelEpisodes)
	optionalAuth.GET("/channels/:channelId/episodes/:episodeId", container.EpisodeHandler.GetEpisode)
	optionalAuth.GET("/recommendations/channels", container.RecommendationHandler.GetRecommendedChannels)
	optionalAuth.GET("/recommendations/episodes", container.RecommendationHandler.GetRecommendedEpisodes)
	optionalAuth.GET("/categories", container.CategoryHandler.ListCategories)
	optionalAuth.POST("/contacts", container.ContactHandler.CreateContact)

	// Admin（認証必須 + 管理者権限必須）
	admin := r.Group("/admin")
	admin.Use(middleware.Auth(container.TokenManager))
	admin.Use(middleware.Admin(container.UserRepository))
	admin.POST("/cleanup/orphaned-media", container.CleanupHandler.CleanupOrphanedMedia)

	// Internal（Cloud Tasks ワーカー用）
	internal := r.Group("/internal")
	internal.Use(middleware.CloudTasksAuth(cfg.GoogleCloudTasksWorkerURL, cfg.GoogleCloudTasksServiceAccountEmail))
	internal.POST("/worker/audio", container.WorkerHandler.ProcessAudioJob)
	internal.POST("/worker/script", container.WorkerHandler.ProcessScriptJob)

	// WebSocket
	r.GET("/ws/jobs", container.WebSocketHandler.HandleJobs)

	return r
}
