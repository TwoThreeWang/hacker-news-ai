package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
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

	// 获取文章的原始内容和评论
	content, err := s.fetchContent(rawStory.URL, rawStory.ID)
	if err != nil {
		return story, fmt.Errorf("获取文章内容失败: %v", err)
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
		Content:     content,
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

// fetchContent 获取文章的原始内容和评论
func (s *HNService) fetchContent(url string, storyID int) (string, error) {
	// 设置请求头
	headers := make(http.Header)
	headers.Set("X-Retain-Images", "none")

	// 获取文章内容
	req, err := http.NewRequest("GET", "https://r.jina.ai/"+url, nil)
	if err != nil {
		return "", fmt.Errorf("创建文章请求失败: %v", err)
	}
	req.Header = headers

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("获取文章内容失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("获取文章失败: %s %s", resp.Status, url)
	}

	articleBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取文章内容失败: %v", err)
	}

	// 获取评论内容
	commentsBody, err := s.fetchComments(storyID)
	if err != nil {
		log.Printf("获取评论内容失败: %v", err)
		// 评论获取失败不影响返回文章内容
		return fmt.Sprintf("\n<article>\n%s\n</article>\n", string(articleBody)), nil
	}

	// 构建返回内容
	var parts []string

	// 添加文章内容
	parts = append(parts, fmt.Sprintf("\n<article>\n%s\n</article>\n", string(articleBody)))

	// 添加评论内容
	parts = append(parts, fmt.Sprintf("\n<comments>\n%s\n</comments>\n", string(commentsBody)))

	// 合并所有内容
	content := strings.Join(parts, "\n---\n")

	// 限制内容长度，避免token过多
	if len(content) > 8000 {
		content = content[:8000]
	}

	return content, nil
}

// fetchComments 获取文章的评论内容
func (s *HNService) fetchComments(storyID int) (string, error) {
	// 获取评论ID列表
	resp, err := s.client.Get(fmt.Sprintf("%s/item/%d.json", s.config.HNAPIBaseURL, storyID))
	if err != nil {
		return "", fmt.Errorf("获取评论列表失败: %v", err)
	}
	defer resp.Body.Close()

	var item struct {
		Kids []int `json:"kids"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return "", fmt.Errorf("解析评论列表失败: %v", err)
	}

	// 获取评论内容和得分
	type commentInfo struct {
		text  string
		by    string
		score int
	}
	var commentsWithScore []commentInfo

	// 获取所有评论的内容和得分
	for _, commentID := range item.Kids {
		commentResp, err := s.client.Get(fmt.Sprintf("%s/item/%d.json", s.config.HNAPIBaseURL, commentID))
		if err != nil {
			continue
		}

		var comment struct {
			Text  string `json:"text"`
			By    string `json:"by"`
			Score int    `json:"score"`
		}
		if err := json.NewDecoder(commentResp.Body).Decode(&comment); err != nil {
			commentResp.Body.Close()
			continue
		}
		commentResp.Body.Close()

		commentsWithScore = append(commentsWithScore, commentInfo{
			text:  comment.Text,
			by:    comment.By,
			score: comment.Score,
		})
	}

	// 按得分排序
	sort.Slice(commentsWithScore, func(i, j int) bool {
		return commentsWithScore[i].score > commentsWithScore[j].score
	})

	// 获取前10条热门评论
	var comments []string
	limit := 10
	if len(commentsWithScore) < limit {
		limit = len(commentsWithScore)
	}

	for i := 0; i < limit; i++ {
		comment := commentsWithScore[i]
		comments = append(comments, fmt.Sprintf("@%s (得分:%d): %s", comment.by, comment.score, comment.text))
	}

	return strings.Join(comments, "\n"), nil
}
