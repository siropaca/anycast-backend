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

// ルーターを設定して返す
func Setup(container *di.Container, cfg *config.Config) *gin.Engine {
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

	// 認証必須のエンドポイント
	authenticated := api.Group("")
	authenticated.Use(middleware.Auth(container.TokenManager))

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

	// Channels
	authenticated.GET("/channels/:channelId", container.ChannelHandler.GetChannel)
	authenticated.POST("/channels", container.ChannelHandler.CreateChannel)
	authenticated.PATCH("/channels/:channelId", container.ChannelHandler.UpdateChannel)
	authenticated.DELETE("/channels/:channelId", container.ChannelHandler.DeleteChannel)
	authenticated.POST("/channels/:channelId/publish", container.ChannelHandler.PublishChannel)
	authenticated.POST("/channels/:channelId/unpublish", container.ChannelHandler.UnpublishChannel)

	// Episodes
	authenticated.POST("/channels/:channelId/episodes", container.EpisodeHandler.CreateEpisode)
	authenticated.PATCH("/channels/:channelId/episodes/:episodeId", container.EpisodeHandler.UpdateEpisode)
	authenticated.DELETE("/channels/:channelId/episodes/:episodeId", container.EpisodeHandler.DeleteEpisode)
	authenticated.POST("/channels/:channelId/episodes/:episodeId/publish", container.EpisodeHandler.PublishEpisode)
	authenticated.POST("/channels/:channelId/episodes/:episodeId/unpublish", container.EpisodeHandler.UnpublishEpisode)
	authenticated.POST("/channels/:channelId/episodes/:episodeId/audio/generate", container.EpisodeHandler.GenerateAudio)

	// Script Lines
	authenticated.GET("/channels/:channelId/episodes/:episodeId/script/lines", container.ScriptLineHandler.ListScriptLines)
	authenticated.POST("/channels/:channelId/episodes/:episodeId/script/lines", container.ScriptLineHandler.CreateScriptLine)
	authenticated.PATCH("/channels/:channelId/episodes/:episodeId/script/lines/:lineId", container.ScriptLineHandler.UpdateScriptLine)
	authenticated.DELETE("/channels/:channelId/episodes/:episodeId/script/lines/:lineId", container.ScriptLineHandler.DeleteScriptLine)
	authenticated.POST("/channels/:channelId/episodes/:episodeId/script/reorder", container.ScriptLineHandler.ReorderScriptLines)

	// Script（台本）
	authenticated.POST("/channels/:channelId/episodes/:episodeId/script/generate", container.ScriptHandler.GenerateScript)
	authenticated.POST("/channels/:channelId/episodes/:episodeId/script/import", container.ScriptHandler.ImportScript)
	authenticated.GET("/channels/:channelId/episodes/:episodeId/script/export", container.ScriptHandler.ExportScript)

	// Voices
	authenticated.GET("/voices", container.VoiceHandler.ListVoices)
	authenticated.GET("/voices/:voiceId", container.VoiceHandler.GetVoice)

	// Categories
	authenticated.GET("/categories", container.CategoryHandler.ListCategories)

	// Images
	authenticated.POST("/images", container.ImageHandler.UploadImage)

	// Admin（認証必須 + 管理者権限必須）
	admin := r.Group("/admin")
	admin.Use(middleware.Auth(container.TokenManager))
	admin.Use(middleware.Admin(container.UserRepository))
	admin.POST("/cleanup/orphaned-media", container.CleanupHandler.CleanupOrphanedMedia)

	return r
}
