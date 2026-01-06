// 孤児メディアファイルをクリーンアップするスクリプト
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"

	"github.com/siropaca/anycast-backend/internal/config"
	"github.com/siropaca/anycast-backend/internal/db"
	"github.com/siropaca/anycast-backend/internal/infrastructure/storage"
	"github.com/siropaca/anycast-backend/internal/repository"
	"github.com/siropaca/anycast-backend/internal/service"
)

func main() {
	dryRun := flag.Bool("dry-run", true, "削除対象の一覧を表示するのみで実際の削除は行わない（デフォルト: true）")
	flag.Parse()

	_ = godotenv.Load() //nolint:errcheck // .env ファイルがなくてもエラーにしない

	cfg := config.Load()

	if cfg.DatabaseURL == "" {
		fmt.Fprintln(os.Stderr, "Error: DATABASE_URL is not set")
		os.Exit(1)
	}

	ctx := context.Background()

	// DB 接続
	dbConn, err := db.New(cfg.DatabaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to connect to database: %v\n", err)
		os.Exit(1)
	}

	// GCS クライアント
	var storageClient storage.Client
	if cfg.GCSBucketName != "" && cfg.GoogleCredentialsJSON != "" {
		storageClient, err = storage.NewGCSClient(ctx, cfg.GCSBucketName, cfg.GoogleCredentialsJSON)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to create storage client: %v\n", err)
		}
	}

	// Repository
	audioRepo := repository.NewAudioRepository(dbConn)
	imageRepo := repository.NewImageRepository(dbConn)

	// Service
	cleanupService := service.NewCleanupService(audioRepo, imageRepo, storageClient)

	// 実行
	fmt.Printf("Cleaning up orphaned media files (dry-run: %v)\n\n", *dryRun)

	result, err := cleanupService.CleanupOrphanedMedia(ctx, *dryRun)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to cleanup orphaned media: %v\n", err)
		os.Exit(1)
	}

	// 結果を表示
	fmt.Printf("Orphaned audios: %d\n", len(result.OrphanedAudios))
	for _, audio := range result.OrphanedAudios {
		fmt.Printf("  - %s (%s, %d bytes)\n", audio.ID, audio.Filename, audio.FileSize)
	}

	fmt.Printf("\nOrphaned images: %d\n", len(result.OrphanedImages))
	for _, image := range result.OrphanedImages {
		fmt.Printf("  - %s (%s, %d bytes)\n", image.ID, image.Filename, image.FileSize)
	}

	if !*dryRun {
		fmt.Printf("\nDeleted audios: %d\n", result.DeletedAudioCount)
		fmt.Printf("Deleted images: %d\n", result.DeletedImageCount)
		fmt.Printf("Failed audios:  %d\n", result.FailedAudioCount)
		fmt.Printf("Failed images:  %d\n", result.FailedImageCount)
	} else {
		fmt.Println("\n(dry-run mode: no files were deleted)")
		fmt.Println("Run with --dry-run=false to actually delete files")
	}
}
