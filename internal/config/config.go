package config

import (
	"os"
)

// 環境を表す型
type Env string

const (
	EnvProduction  Env = "production"
	EnvDevelopment Env = "development"
)

// アプリケーション設定
type Config struct {
	Port        string
	DatabaseURL string
	AppEnv      Env
}

// 環境変数から設定を読み込む
func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8081"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		AppEnv:      Env(getEnv("APP_ENV", string(EnvDevelopment))),
	}
}

// 環境変数を取得し、未設定の場合はデフォルト値を返す
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
