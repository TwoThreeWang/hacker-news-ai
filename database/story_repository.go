package database

import (
	"fmt"
	"time"

	"github.com/hacker-news-ai/models"
	"gorm.io/gorm"
)

// StoryRepository 文章数据库操作封装
type StoryRepository struct {
	db *gorm.DB
}

// NewStoryRepository 创建文章数据库操作实例
func NewStoryRepository() *StoryRepository {
	return &StoryRepository{
		db: GetDB(),
	}
}

// SaveStories 保存文章列表和博客内容到数据库
func (r *StoryRepository) SaveStories(blogContent, title, Pid string) error {
	// 开启事务
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	// 创建新的TbPost记录
	post := models.TbPost{
		Title:        title,
		Content:      blogContent,
		Status:       "Active",
		CreatedAt:    time.Now(),
		UpVote:       0,
		CollectVote:  0,
		Type:         "ask",
		UserID:       1,
		Pid:          Pid,
		CommentCount: 0,
		Point:        0.1,
		Top:          0,
		ClickVote:    0,
	}

	// 插入文章并获取ID
	if err := tx.Create(&post).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("保存文章失败: %v", err)
	}

	// 更新用户文章计数
	if err := tx.Model(&models.TbUser{}).Where("id = ?", 1).UpdateColumn("\"postCount\"", gorm.Expr("\"postCount\" + ?", 1)).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新用户文章计数失败: %v", err)
	}

	// 插入文章标签关联
	postTag := models.TbPostTag{
		TbPostID: post.ID,
		TbTagID:  15,
	}
	if err := tx.Create(&postTag).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("保存文章标签关联失败: %v", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}
