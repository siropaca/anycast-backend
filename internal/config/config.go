package config

import "os"

// Config はアプリケーション設定
type Config struct {
	Port        string
	DatabaseURL string
	AppEnv      string
}

// Load は環境変数から設定を読み込む
func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8081"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		AppEnv:      getEnv("APP_ENV", "development"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
