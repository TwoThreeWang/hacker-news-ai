package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/hacker-news-ai/config"
	"github.com/hacker-news-ai/models"
	"google.golang.org/api/option"
)

type AIService struct {
	config  *config.Config
	client  *http.Client
	service *genai.Client
}

func NewAIService(cfg *config.Config) (*AIService, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(cfg.GeminiAPIKey))
	if err != nil {
		return nil, fmt.Errorf("初始化Gemini客户端失败: %v", err)
	}

	return &AIService{
		config: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		service: client,
	}, nil
}

// GenerateSummary 为文章生成中文总结
func (s *AIService) GenerateSummary(story *models.Story) error {
	// 获取文章内容
	content, err := s.fetchContent(story.URL, story.ID)
	if err != nil {
		return fmt.Errorf("获取文章内容失败: %v", err)
	}

	// 构建提示词
	prompt := fmt.Sprintf(
		`你是 Hacker News 中文博客的编辑助理，擅长将 Hacker News 上的文章和评论整理成引人入胜的博客内容。内容受众主要为软件开发者和科技爱好者。

【工作目标】
- 接收并阅读来自 Hacker News 的文章与评论。
- 先简明介绍文章的主要话题，再对其要点进行精炼说明。
- 分析并总结评论区的不同观点，展现多样化视角。
- 以清晰直接的口吻进行讨论，像与朋友交谈般简洁易懂。
- 按照逻辑顺序，使用二级标题 (如"## 标题") 与分段正文形式呈现播客的核心精简内容。
- 所有违反中国大陆法律和政治立场的内容，都跳过。

【输出要求】
- 直接输出正文，不要返回前言。
- 直接进入主要内容的总结与讨论：
  * 第 1-2 句：概括适合搜索引擎收录的文章主题，主题需要使用二级标题。
  * 第 3-15 句：详细阐述文章的重点内容。
  * 第 16-25 句：总结和对评论观点的分析，体现多角度探讨。
- 直接返回 Markdown 格式的正文内容。
- 换行不要使用\n,使用两个回车。`,
		story.Title,
		content,
	)

	// 调用Gemini API生成总结
	ctx := context.Background()
	model := s.service.GenerativeModel("gemini-2.0-flash")
	model.SetTemperature(0.3)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return fmt.Errorf("生成总结失败: %v", err)
	}

	if len(resp.Candidates) == 0 {
		return fmt.Errorf("未能生成有效的总结")
	}

	// 更新文章的总结信息
	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		content := resp.Candidates[0].Content.Parts[0]
		story.Summary = fmt.Sprintf("%s", content)
		return nil
	} else {
		return fmt.Errorf("未能获取到有效的总结内容")
	}
}

// fetchComments 获取文章的评论内容
func (s *AIService) fetchComments(storyID int) (string, error) {
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

// fetchContent 获取文章的原始内容和评论
func (s *AIService) fetchContent(url string, storyID int) (string, error) {
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
