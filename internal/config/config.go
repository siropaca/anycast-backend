package config

import (
	"os"
	"strings"
)

// Env は環境を表す型
type Env string

const (
	EnvProduction  Env = "production"
	EnvDevelopment Env = "development"
)

// DBLogLevel はデータベースのログレベルを表す型
type DBLogLevel string

const (
	DBLogLevelSilent DBLogLevel = "silent"
	DBLogLevelError  DBLogLevel = "error"
	DBLogLevelWarn   DBLogLevel = "warn"
	DBLogLevelInfo   DBLogLevel = "info"
)

// Config はアプリケーション設定
type Config struct {
	Port                                string
	DatabaseURL                         string
	DBLogLevel                          DBLogLevel
	AppEnv                              Env
	AuthSecret                          string
	CORSAllowedOrigins                  []string
	OpenAIAPIKey                        string
	GoogleCloudProjectID                string
	GoogleCloudCredentialsJSON          string
	GoogleCloudStorageBucketName        string
	GoogleCloudTasksLocation            string
	GoogleCloudTasksQueueName           string
	GoogleCloudTasksServiceAccountEmail string
	GoogleCloudTasksWorkerURL           string
	// Gemini TTS の location（デフォルト: us-central1）
	GoogleCloudTTSLocation string
	// Claude API キー
	ClaudeAPIKey string
	// Gemini LLM のロケーション（デフォルト: asia-northeast1）
	GeminiLLMLocation string
	// Slack Webhook URL（空の場合は通知無効）
	SlackWebhookURL string
}

// Load は環境変数から設定を読み込む
func Load() *Config {
	return &Config{
		Port:                                getEnv("PORT", "8081"),
		DatabaseURL:                         getEnv("DATABASE_URL", ""),
		DBLogLevel:                          DBLogLevel(getEnv("DB_LOG_LEVEL", string(DBLogLevelSilent))),
		AppEnv:                              Env(getEnv("APP_ENV", string(EnvDevelopment))),
		AuthSecret:                          getEnv("AUTH_SECRET", ""),
		CORSAllowedOrigins:                  getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3210"}),
		OpenAIAPIKey:                        getEnv("OPENAI_API_KEY", ""),
		GoogleCloudProjectID:                getEnv("GOOGLE_CLOUD_PROJECT_ID", ""),
		GoogleCloudCredentialsJSON:          getEnv("GOOGLE_CLOUD_CREDENTIALS_JSON", ""),
		GoogleCloudStorageBucketName:        getEnv("GOOGLE_CLOUD_STORAGE_BUCKET_NAME", ""),
		GoogleCloudTasksLocation:            getEnv("GOOGLE_CLOUD_TASKS_LOCATION", "asia-northeast1"),
		GoogleCloudTasksQueueName:           getEnv("GOOGLE_CLOUD_TASKS_QUEUE_NAME", "audio-generation-queue"),
		GoogleCloudTasksServiceAccountEmail: getEnv("GOOGLE_CLOUD_TASKS_SERVICE_ACCOUNT_EMAIL", ""),
		GoogleCloudTasksWorkerURL:           getEnv("GOOGLE_CLOUD_TASKS_WORKER_URL", ""),
		GoogleCloudTTSLocation:              getEnv("GOOGLE_CLOUD_TTS_LOCATION", "us-central1"),
		ClaudeAPIKey:                        getEnv("CLAUDE_API_KEY", ""),
		GeminiLLMLocation:                   getEnv("GEMINI_LLM_LOCATION", "asia-northeast1"),
		SlackWebhookURL:                     getEnv("SLACK_WEBHOOK_URL", ""),
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
