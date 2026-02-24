package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
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

	// グレースフルシャットダウンのタイムアウト
	shutdownTimeout = 30 * time.Second
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
	log := logger.Default()

	// DB 初期化
	database, err := db.New(cfg.DatabaseURL, cfg.DBLogLevel)
	if err != nil {
		log.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	// DI コンテナ構築
	ctx := context.Background()
	container := di.NewContainer(ctx, database, cfg)

	// ルーター設定
	r := router.Setup(container, cfg)

	// サーバー起動
	if cfg.AppEnv == config.EnvDevelopment {
		log.Info("Server listening on http://localhost:" + cfg.Port)
		log.Info("Swagger UI: http://localhost:" + cfg.Port + "/swagger/index.html")
	}

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	// シグナルハンドリング用のコンテキスト
	srvCtx, srvStop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)

	// サーバーをゴルーチンで起動
	srvErr := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			srvErr <- err
		}
		close(srvErr)
	}()

	// シグナル受信またはサーバーエラーを待機
	select {
	case err := <-srvErr:
		srvStop()
		if err != nil {
			var opErr *net.OpError
			if errors.As(err, &opErr) {
				var syscallErr *os.SyscallError
				if errors.As(opErr.Err, &syscallErr) && errors.Is(syscallErr.Err, syscall.EADDRINUSE) {
					log.Error("Port "+cfg.Port+" is already in use. Please stop the existing process or use a different port.", "error", err)
					os.Exit(1)
				}
			}
			log.Error("Failed to start server", "error", err)
			os.Exit(1)
		}

	case <-srvCtx.Done():
		srvStop()
		log.Info("Shutdown signal received")
	}

	// グレースフルシャットダウン
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)

	log.Info("Shutting down server...")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("Server shutdown error", "error", err)
	}
	shutdownCancel()

	// リソースのクリーンアップ
	if err := container.Close(); err != nil {
		log.Error("Container close error", "error", err)
	}

	sqlDB, err := database.DB()
	if err == nil {
		if err := sqlDB.Close(); err != nil {
			log.Error("Database close error", "error", err)
		}
	}

	log.Info("Server stopped")
}
