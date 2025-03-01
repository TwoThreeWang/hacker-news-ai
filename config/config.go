package config

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type Config struct {
	// Google Gemini API配置
	GeminiAPIKey string `json:"gemini_api_key"`

	// Hacker News API配置
	HNAPIBaseURL string `json:"hn_api_base_url"`
	// Dev.to API配置
	DevAPIBaseURL string `json:"dev_api_base_url"`
	// 每日获取的热门文章数量
	TopStoriesLimit int `json:"top_stories_limit"`
	// 抓取间隔（分钟）
	FetchInterval int `json:"fetch_interval"`

	// 数据库配置
	DBHost     string `json:"db_host"`
	DBPort     int    `json:"db_port"`
	DBUser     string `json:"db_user"`
	DBPassword string `json:"db_password"`
	DBName     string `json:"db_name"`
}

var (
	config *Config
	once   sync.Once
)

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*Config, error) {
	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	once.Do(func() {
		// 初始化默认配置
		config = &Config{
			HNAPIBaseURL:    "https://hacker-news.firebaseio.com/v0",
			DevAPIBaseURL:   "https://dev.to/api",
			TopStoriesLimit: 30,
			FetchInterval:   60,
		}

		// 解析JSON配置文件
		if err := json.Unmarshal(data, config); err != nil {
			fmt.Printf("解析配置文件失败: %v，将使用默认配置", err)
		}
	})

	return config, nil
}

// GetConfig 获取配置实例
func GetConfig() *Config {
	return config
}
