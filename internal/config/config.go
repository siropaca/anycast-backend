package config

import (
	"os"
	"strings"
)

// 環境を表す型
type Env string

const (
	EnvProduction  Env = "production"
	EnvDevelopment Env = "development"
)

// アプリケーション設定
type Config struct {
	Port               string
	DatabaseURL        string
	AppEnv             Env
	AuthSecret         string
	CORSAllowedOrigins []string
}

// 環境変数から設定を読み込む
func Load() *Config {
	return &Config{
		Port:               getEnv("PORT", "8081"),
		DatabaseURL:        getEnv("DATABASE_URL", ""),
		AppEnv:             Env(getEnv("APP_ENV", string(EnvDevelopment))),
		AuthSecret:         getEnv("AUTH_SECRET", ""),
		CORSAllowedOrigins: getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3210"}),
	}
}

// 環境変数を取得し、未設定の場合はデフォルト値を返す
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}

// 環境変数をカンマ区切りで分割してスライスとして取得する
func getEnvAsSlice(key string, defaultValue []string) []string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	var result []string
	for _, v := range strings.Split(value, ",") {
		trimmed := strings.TrimSpace(v)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}
