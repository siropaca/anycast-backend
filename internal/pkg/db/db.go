package db

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/siropaca/anycast-backend/internal/config"
)

// New はデータベース接続を作成して返す
//
// @param databaseURL - データベース接続 URL
// @param dbLogLevel - ログレベル（silent, error, warn, info）
func New(databaseURL string, dbLogLevel config.DBLogLevel) (*gorm.DB, error) {
	logLevel := toGormLogLevel(dbLogLevel)

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	}

	db, err := gorm.Open(postgres.Open(databaseURL), gormConfig)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// toGormLogLevel は config.DBLogLevel を gorm の logger.LogLevel に変換する
func toGormLogLevel(level config.DBLogLevel) logger.LogLevel {
	switch level {
	case config.DBLogLevelSilent:
		return logger.Silent
	case config.DBLogLevelError:
		return logger.Error
	case config.DBLogLevelWarn:
		return logger.Warn
	case config.DBLogLevelInfo:
		return logger.Info
	default:
		return logger.Info
	}
}
