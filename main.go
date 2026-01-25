package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/siropaca/anycast-backend/internal/config"
	"github.com/siropaca/anycast-backend/internal/di"
	"github.com/siropaca/anycast-backend/internal/pkg/db"
	"github.com/siropaca/anycast-backend/internal/pkg/logger"
	"github.com/siropaca/anycast-backend/internal/router"
)

const (
	// HTTP サーバーのタイムアウト設定
	readTimeout  = 10 * time.Second
	writeTimeout = 180 * time.Second // LLM 生成を考慮して長めに設定
	idleTimeout  = 60 * time.Second
)

// @title Anycast API
// @version 1.0
// @description AI ポッドキャスト作成・配信プラットフォーム API

// @host localhost:8081
// @BasePath /api/v1
func main() {
	_ = godotenv.Load() //nolint:errcheck // .env ファイルがなくてもエラーにしない

	// 設定読み込み
	cfg := config.Load()

	// Logger 初期化
	logger.Init(cfg.AppEnv)

	// DB 初期化
	database, err := db.New(cfg.DatabaseURL, cfg.AppEnv)
	if err != nil {
		logger.Default().Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	// DI コンテナ構築
	ctx := context.Background()
	container := di.NewContainer(ctx, database, cfg)

	// ルーター設定
	r := router.Setup(container, cfg)

	// サーバー起動
	if cfg.AppEnv == config.EnvDevelopment {
		logger.Default().Info("Server listening on http://localhost:" + cfg.Port)
		logger.Default().Info("Swagger UI: http://localhost:" + cfg.Port + "/swagger/index.html")
	}

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		var opErr *net.OpError
		if errors.As(err, &opErr) {
			var syscallErr *os.SyscallError
			if errors.As(opErr.Err, &syscallErr) && errors.Is(syscallErr.Err, syscall.EADDRINUSE) {
				logger.Default().Error("Port "+cfg.Port+" is already in use. Please stop the existing process or use a different port.", "error", err)
				os.Exit(1)
			}
		}

		logger.Default().Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}
