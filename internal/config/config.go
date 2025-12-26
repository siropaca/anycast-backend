package config

import "os"

// アプリケーション設定
type Config struct {
	Port        string
	DatabaseURL string
	AppEnv      string
}

// 環境変数から設定を読み込む
func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8081"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		AppEnv:      getEnv("APP_ENV", "development"),
	}
}

// getEnv は環境変数を取得し、未設定の場合はデフォルト値を返す
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
