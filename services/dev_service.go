package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/hacker-news-ai/config"
	"github.com/hacker-news-ai/models"
)

type DevService struct {
	config *config.Config
	client *http.Client
}

func NewDevService(cfg *config.Config) *DevService {
	return &DevService{
		config: cfg,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// FetchTopStories 获取dev.to热门文章列表
func (s *DevService) FetchTopStories() ([]models.Story, error) {
	// 获取热门文章列表
	resp, err := s.client.Get(fmt.Sprintf("%s/articles?top=1d&per_page=%d", s.config.DevAPIBaseURL, s.config.TopStoriesLimit))
	if err != nil {
		return nil, fmt.Errorf("获取dev.to热门文章列表失败: %v", err)
	}
	defer resp.Body.Close()

	// 解析API响应
	var articles []struct {
		ID          int       `json:"id"`
		Title       string    `json:"title"`
		URL         string    `json:"url"`
		PublishedAt time.Time `json:"published_at"`
		User        struct {
			Username string `json:"username"`
		} `json:"user"`
		PositiveReactionsCount int `json:"positive_reactions_count"`
		CommentsCount          int `json:"comments_count"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&articles); err != nil {
		return nil, fmt.Errorf("解析dev.to文章列表失败: %v", err)
	}

	// 转换为Story模型
	var stories []models.Story
	for _, article := range articles {
		// 获取文章的详细内容
		story, err := s.FetchStory(article.ID)
		if err != nil {
			continue
		}
		stories = append(stories, story)
	}

	return stories, nil
}

// FetchStory 获取单个文章的详细信息
func (s *DevService) FetchStory(id int) (models.Story, error) {
	var story models.Story

	// 获取文章详情
	resp, err := s.client.Get(fmt.Sprintf("%s/articles/%d", s.config.DevAPIBaseURL, id))
	if err != nil {
		return story, fmt.Errorf("获取dev.to文章详情失败: %v", err)
	}
	defer resp.Body.Close()

	// 解析API响应
	var article struct {
		ID          int       `json:"id"`
		Title       string    `json:"title"`
		URL         string    `json:"url"`
		PublishedAt time.Time `json:"published_at"`
		User        struct {
			Username string `json:"username"`
		} `json:"user"`
		BodyMarkdown           string `json:"body_markdown"`
		PositiveReactionsCount int    `json:"positive_reactions_count"`
		CommentsCount          int    `json:"comments_count"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&article); err != nil {
		return story, fmt.Errorf("解析dev.to文章详情失败: %v", err)
	}

	// 拼接文章的原始内容和评论
	content, err := s.fetchContent(article.BodyMarkdown, article.ID)
	if err != nil {
		return story, fmt.Errorf("获取文章内容失败: %v", err)
	}

	// 转换为Story模型
	story = models.Story{
		ID:          article.ID,
		Title:       article.Title,
		URL:         article.URL,
		Score:       article.PositiveReactionsCount,
		Time:        article.PublishedAt,
		By:          article.User.Username,
		Descendants: article.CommentsCount,
		Content:     content,
	}

	return story, nil
}

// fetchContent 获取文章的原始内容和评论
func (s *DevService) fetchContent(content string, articleID int) (string, error) {
	// 获取评论内容
	commentsBody, err := s.fetchComments(articleID)
	if err != nil {
		log.Printf("获取评论内容失败: %v", err)
		// 评论获取失败不影响返回文章内容
		return fmt.Sprintf("\n<article>\n%s\n</article>\n", content), nil
	}

	// 构建返回内容
	var parts []string

	// 添加文章内容
	parts = append(parts, fmt.Sprintf("\n<article>\n%s\n</article>\n", content))

	// 添加评论内容
	parts = append(parts, fmt.Sprintf("\n<comments>\n%s\n</comments>\n", commentsBody))

	// 合并所有内容
	fullContent := strings.Join(parts, "\n---\n")

	// 限制内容长度，避免token过多
	if len(fullContent) > 8000 {
		fullContent = fullContent[:8000]
	}

	return fullContent, nil
}

// fetchComments 获取文章的评论内容
func (s *DevService) fetchComments(articleID int) (string, error) {
	// 获取评论列表
	resp, err := s.client.Get(fmt.Sprintf("%s/comments?a_id=%d&order=popular", s.config.DevAPIBaseURL, articleID))
	if err != nil {
		return "", fmt.Errorf("获取dev.to评论列表失败: %v", err)
	}
	defer resp.Body.Close()

	// 解析API响应
	var comments []struct {
		BodyMarkdown string `json:"body_markdown"`
		User         struct {
			Username string `json:"username"`
		} `json:"user"`
		PositiveReactionsCount int `json:"positive_reactions_count"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&comments); err != nil {
		return "", fmt.Errorf("解析dev.to评论列表失败: %v", err)
	}

	// 获取前10条热门评论
	var commentStrings []string
	limit := 10
	if len(comments) < limit {
		limit = len(comments)
	}

	for i := 0; i < limit; i++ {
		comment := comments[i]
		if comment.BodyMarkdown != "" {
			commentStrings = append(commentStrings, fmt.Sprintf("@%s (得分:%d): %s", comment.User.Username, comment.PositiveReactionsCount, comment.BodyMarkdown))
		}
	}

	return strings.Join(commentStrings, "\n"), nil
}
