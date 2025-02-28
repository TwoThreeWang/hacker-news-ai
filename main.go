package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hacker-news-ai/config"
	"github.com/hacker-news-ai/database"
	"github.com/hacker-news-ai/services"
)

func main() {
	// 设置日志输出
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// 初始化配置
	cfg, err := config.LoadConfig("config/config.json")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化数据库
	if err := database.InitDB(cfg); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	// 初始化数据库操作服务
	storyRepo := database.NewStoryRepository()

	// 初始化服务
	hnService := services.NewHNService(cfg)
	aiService, err := services.NewAIService(cfg)
	if err != nil {
		log.Fatalf("初始化AI服务失败: %v", err)
	}

	fmt.Println("Hacker News AI 助手启动于:", time.Now().Format("2006-01-02 15:04:05"))

	// 立即执行一次
	fetchAndProcessStories(hnService, aiService, storyRepo)
}

// fetchAndProcessStories 获取并处理热门文章
func fetchAndProcessStories(hnService *services.HNService, aiService *services.AIService, storyRepo *database.StoryRepository) {
	// 获取热门文章
	stories, err := hnService.FetchTopStories()
	if err != nil {
		log.Printf("获取热门文章失败: %v", err)
		return
	}
	// 打印获取到的文章数量
	fmt.Printf("获取到 %d 篇文章\n", len(stories))
	blogContent := ""
	// 为每篇文章生成中文总结
	for i := range stories {
		fmt.Printf("%d. %s\n", i, stories[i].Title)
		if err := aiService.GenerateSummary(&stories[i]); err != nil {
			log.Printf("生成文章总结失败 [%s]: %v", stories[i].Title, err)
			continue
		}

		content := fmt.Sprintf("%s\n\n- 原文: [%s](%s)\n- 作者: %s\n- 评分: %d\n- 评论数: %d\n- 发布时间: %s\n\n---\n\n",
			stories[i].Summary,
			stories[i].Title,
			stories[i].URL,
			stories[i].By,
			stories[i].Score,
			stories[i].Descendants,
			stories[i].Time.Format("2006-01-02 15:04:05"),
		)
		blogContent += content
	}

	// 保存文章和博客内容到数据库
	if err := storyRepo.SaveStories(blogContent); err != nil {
		log.Printf("保存数据到数据库失败: %v", err)
	}

	//fmt.Println(blogContent)
	fmt.Println("Hacker News AI 助手运行完成于:", time.Now().Format("2006-01-02 15:04:05"))
}
