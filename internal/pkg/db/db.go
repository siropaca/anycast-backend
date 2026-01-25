package db

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/siropaca/anycast-backend/internal/config"
)

// New はデータベース接続を作成して返す
//
// @param databaseURL - データベース接続URL
// @param env - 実行環境
// @returns データベース接続
func New(databaseURL string, env config.Env) (*gorm.DB, error) {
	var logLevel logger.LogLevel
	if env == config.EnvProduction {
		logLevel = logger.Error
	} else {
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, err
	}

	return db, nil
}
