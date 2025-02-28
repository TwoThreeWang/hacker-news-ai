package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hacker-news-ai/config"
	"github.com/hacker-news-ai/models"
)

type HNService struct {
	config *config.Config
	client *http.Client
}

func NewHNService(cfg *config.Config) *HNService {
	return &HNService{
		config: cfg,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// FetchTopStories 获取热门文章列表
func (s *HNService) FetchTopStories() ([]models.Story, error) {
	// 获取热门文章ID列表
	resp, err := s.client.Get(fmt.Sprintf("%s/topstories.json", s.config.HNAPIBaseURL))
	if err != nil {
		return nil, fmt.Errorf("获取热门文章列表失败: %v", err)
	}
	defer resp.Body.Close()

	var storyIDs []int
	if err := json.NewDecoder(resp.Body).Decode(&storyIDs); err != nil {
		return nil, fmt.Errorf("解析热门文章列表失败: %v", err)
	}

	// 限制获取的文章数量
	if len(storyIDs) > s.config.TopStoriesLimit {
		storyIDs = storyIDs[:s.config.TopStoriesLimit]
	}

	// 获取每个文章的详细信息
	var stories []models.Story
	for _, id := range storyIDs {
		story, err := s.FetchStory(id)
		if err != nil {
			continue
		}
		stories = append(stories, story)
	}

	return stories, nil
}

// FetchStory 获取单个文章的详细信息
func (s *HNService) FetchStory(id int) (models.Story, error) {
	var story models.Story

	resp, err := s.client.Get(fmt.Sprintf("%s/item/%d.json", s.config.HNAPIBaseURL, id))
	if err != nil {
		return story, fmt.Errorf("获取文章详情失败: %v", err)
	}
	defer resp.Body.Close()

	// 解析API响应
	var rawStory struct {
		ID          int    `json:"id"`
		Title       string `json:"title"`
		URL         string `json:"url"`
		Score       int    `json:"score"`
		Time        int64  `json:"time"`
		By          string `json:"by"`
		Descendants int    `json:"descendants"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&rawStory); err != nil {
		return story, fmt.Errorf("解析文章详情失败: %v", err)
	}

	// 转换为Story模型
	story = models.Story{
		ID:          rawStory.ID,
		Title:       rawStory.Title,
		URL:         getStoryURL(rawStory.URL, rawStory.ID),
		Score:       rawStory.Score,
		Time:        time.Unix(rawStory.Time, 0),
		By:          rawStory.By,
		Descendants: rawStory.Descendants,
	}

	return story, nil
}

// getStoryURL 获取文章URL，如果原始URL为空则使用HN默认链接
func getStoryURL(originalURL string, storyID int) string {
	if originalURL != "" {
		return originalURL
	}
	return fmt.Sprintf("https://news.ycombinator.com/item?id=%d", storyID)
}
