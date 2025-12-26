package main

import (
	"log"

	"github.com/joho/godotenv"

	"github.com/siropaca/anycast-backend/internal/config"
	"github.com/siropaca/anycast-backend/internal/db"
	"github.com/siropaca/anycast-backend/internal/di"
	"github.com/siropaca/anycast-backend/internal/logger"
	"github.com/siropaca/anycast-backend/internal/router"
)

// @title Anycast API
// @version 1.0
// @description AI ポッドキャスト作成・配信プラットフォーム API

// @host localhost:8081
// @BasePath /api/v1
func main() {
	godotenv.Load()

	// 設定読み込み
	cfg := config.Load()

	// Logger 初期化
	logger.Init(cfg.AppEnv)

	// DB 初期化
	database, err := db.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// DI コンテナ構築
	container := di.NewContainer(database)

	// ルーター設定
	r := router.Setup(container, cfg)

	// サーバー起動
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
