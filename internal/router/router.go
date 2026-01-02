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
	authenticated.GET("/me/channels", container.ChannelHandler.ListMyChannels)

	// Channels
	authenticated.GET("/channels/:channelId", container.ChannelHandler.GetChannel)
	authenticated.POST("/channels", container.ChannelHandler.CreateChannel)
	authenticated.PATCH("/channels/:channelId", container.ChannelHandler.UpdateChannel)
	authenticated.DELETE("/channels/:channelId", container.ChannelHandler.DeleteChannel)

	// Voices
	authenticated.GET("/voices", container.VoiceHandler.ListVoices)
	authenticated.GET("/voices/:voiceId", container.VoiceHandler.GetVoice)

	// Categories
	authenticated.GET("/categories", container.CategoryHandler.ListCategories)

	return r
}
