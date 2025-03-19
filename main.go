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

	fmt.Println("AI 总结助手启动于:", time.Now().Format("2006-01-02 15:04:05"))

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
	devService := services.NewDevService(cfg)
	aiService, err := services.NewAIService(cfg)
	if err != nil {
		log.Fatalf("初始化AI服务失败: %v", err)
	}

	// Hack News AI 助手运行
	fetchAndProcessHN(hnService, aiService, storyRepo)
	// dev.to AI 助手运行
	fetchAndProcessDev(devService, aiService, storyRepo)

	fmt.Println("AI 总结助手结束于:", time.Now().Format("2006-01-02 15:04:05"))
}

// fetchAndProcessHN 获取并处理HN热门文章
func fetchAndProcessHN(hnService *services.HNService, aiService *services.AIService, storyRepo *database.StoryRepository) {
	fmt.Println("Hacker News AI 助手启动于:", time.Now().Format("2006-01-02 15:04:05"))
	// 获取热门文章
	stories, err := hnService.FetchTopStories()
	if err != nil {
		log.Printf("获取热门文章失败: %v", err)
		return
	}
	// 打印获取到的文章数量
	fmt.Printf("Hacker News 获取到 %d 篇文章\n", len(stories))
	blogContent := ""
	// 为每篇文章生成中文总结
	for i := range stories {
		fmt.Printf("%d. %s\n", i, stories[i].Title)
		if err := aiService.GenerateSummary(&stories[i]); err != nil {
			log.Printf("Hacker News 生成文章总结失败 [%s]: %v", stories[i].Title, err)
			continue
		}
		hnUrl := fmt.Sprintf("https://news.ycombinator.com/item?id=%d", stories[i].ID)

		content := fmt.Sprintf("%s\n\n- 原文: [%s](%s)\n- Hacker News: [%s](%s)\n- 作者: %s\n- 评分: %d\n- 评论数: %d\n- 发布时间: %s\n\n---\n\n",
			stories[i].Summary,
			stories[i].Title,
			stories[i].URL,
			hnUrl,
			hnUrl,
			stories[i].By,
			stories[i].Score,
			stories[i].Descendants,
			stories[i].Time.Format("2006-01-02 15:04:05"),
		)
		blogContent += content
		// 每次分析间隔3秒，防止api频率限制
		time.Sleep(3 * time.Second)
	}
	if blogContent == "" {
		fmt.Println("Hacker News AI 助手运行错误:", time.Now().Format("2006-01-02 15:04:05"))
		return
	}
	today := time.Now().Format("20060102")
	blogContent = fmt.Sprintf("## Hacker News 中文精选 NO.%s\n\n一个基于 Hacker News 的中文日报项目，每天自动抓取 Hacker News 热门文章及评论，通过 AI 生成中文解读与总结，传递科技前沿信息。\n\n![Hacker News 中文精选](https://cdn.wangtwothree.com/imgur/f6uVgbS.jpeg)\n---\n\n%s", today, blogContent)
	title := fmt.Sprintf("每日科技新知 NO.%s：Hacker News 中文解读，科技前沿热点速递", today)
	Pid := fmt.Sprintf("HN%s", today)
	// 保存文章和博客内容到数据库
	if err := storyRepo.SaveStories(blogContent, title, Pid); err != nil {
		log.Printf("保存数据到数据库失败: %v", err)
	}

	//fmt.Println(blogContent)
	fmt.Println("Hacker News AI 助手运行完成于:", time.Now().Format("2006-01-02 15:04:05"))
}

// fetchAndProcessDev 获取并处理dev.to热门文章
func fetchAndProcessDev(devService *services.DevService, aiService *services.AIService, storyRepo *database.StoryRepository) {
	fmt.Println("dev.to AI 助手启动于:", time.Now().Format("2006-01-02 15:04:05"))
	// 获取热门文章
	stories, err := devService.FetchTopStories()
	if err != nil {
		log.Printf("获取dev.to热门文章失败: %v", err)
		return
	}
	// 打印获取到的文章数量
	fmt.Printf("dev.to 获取到 %d 篇dev.to文章\n", len(stories))
	blogContent := ""
	// 为每篇文章生成中文总结
	for i := range stories {
		fmt.Printf("%d. %s\n", i, stories[i].Title)
		if err := aiService.GenerateSummary(&stories[i]); err != nil {
			log.Printf("dev.to 生成文章总结失败 [%s]: %v", stories[i].Title, err)
			continue
		}

		content := fmt.Sprintf("%s\n\n- 原文: [%s](%s)\n- 作者: %s\n- 点赞数: %d\n- 评论数: %d\n- 发布时间: %s\n\n---\n\n",
			stories[i].Summary,
			stories[i].Title,
			stories[i].URL,
			stories[i].By,
			stories[i].Score,
			stories[i].Descendants,
			stories[i].Time.Format("2006-01-02 15:04:05"),
		)
		blogContent += content
		// 每次分析间隔3秒，防止api频率限制
		time.Sleep(3 * time.Second)
	}
	if blogContent == "" {
		fmt.Println("dev.to AI 助手运行错误:", time.Now().Format("2006-01-02 15:04:05"))
		return
	}
	today := time.Now().Format("20060102")
	blogContent = fmt.Sprintf("## DEV 社区中文精选 NO.%s\n\nDev Community 是一个面向全球开发者的技术博客与协作平台，本文是基于 dev.to 的中文日报项目，每天自动抓取 Dev Community 热门文章及评论，通过 AI 生成中文解读与总结，传递科技前沿信息。\n\n![Dev Community 中文精选](https://cdn.wangtwothree.com/imgur/ebLSg8b.png)\n---\n\n%s", today, blogContent)
	title := fmt.Sprintf("开发者简报 NO.%s：DEV 社区中文解读，全球开发者技术瞭望", today)
	Pid := fmt.Sprintf("DEV%s", today)
	// 保存文章和博客内容到数据库
	if err := storyRepo.SaveStories(blogContent, title, Pid); err != nil {
		log.Printf("保存数据到数据库失败: %v", err)
	}

	fmt.Println("dev.to AI 助手运行完成于:", time.Now().Format("2006-01-02 15:04:05"))
}
