package database

import (
	"fmt"
	"sync"

	"github.com/hacker-news-ai/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	db   *gorm.DB
	once sync.Once
)

// InitDB 初始化数据库连接
func InitDB(cfg *config.Config) error {
	var err error
	once.Do(func() {
		// 构建数据库连接字符串
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
			cfg.DBHost,
			cfg.DBUser,
			cfg.DBPassword,
			cfg.DBName,
			cfg.DBPort,
		)

		// 连接数据库
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			return
		}
	})

	return err
}

// GetDB 获取数据库连接实例
func GetDB() *gorm.DB {
	return db
}
