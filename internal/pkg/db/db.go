package db

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/siropaca/anycast-backend/internal/config"
)

// New はデータベース接続を作成して返す
//
// databaseURL はデータベース接続 URL、env は実行環境を指定する
func New(databaseURL string, env config.Env) (*gorm.DB, error) {
	var gormConfig *gorm.Config

	if env == config.EnvProduction {
		gormConfig = &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		}
	} else {
		gormConfig = &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		}
	}

	db, err := gorm.Open(postgres.Open(databaseURL), gormConfig)
	if err != nil {
		return nil, err
	}

	return db, nil
}
