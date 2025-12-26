package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/siropaca/anycast-backend/internal/config"
	"github.com/siropaca/anycast-backend/internal/di"
	"github.com/siropaca/anycast-backend/internal/logger"
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

	// Swagger（本番環境では無効）
	if cfg.AppEnv != logger.EnvProduction {
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
	{
		// Voices
		api.GET("/voices", container.VoiceHandler.ListVoices)
		api.GET("/voices/:voiceId", container.VoiceHandler.GetVoice)
	}

	return r
}
