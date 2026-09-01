package controllers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"exchangeapp/config"
	"exchangeapp/global"
	"exchangeapp/models"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"gorm.io/gorm"
)

// ──────────────────────────────────────────────
// DeepSeek API 请求 / 响应结构
// ──────────────────────────────────────────────

type dsMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type dsRequest struct {
	Model    string      `json:"model"`
	Messages []dsMessage `json:"messages"`
	Stream   bool        `json:"stream"`
}

// SSE chunk 中的 delta
type dsDelta struct {
	Content string `json:"content"`
}

type dsChoice struct {
	Delta        dsDelta `json:"delta"`
	FinishReason string  `json:"finish_reason"`
}

type dsChunk struct {
	Choices []dsChoice `json:"choices"`
}

// ──────────────────────────────────────────────
// 语言检测（简单启发：中文字符占比）
// ──────────────────────────────────────────────

func detectLanguage(text string) string {
	if len(text) == 0 {
		return "unknown"
	}
	chineseCount := 0
	total := 0
	for _, r := range text {
		total++
		if r >= 0x4E00 && r <= 0x9FFF {
			chineseCount++
		}
	}
	if total == 0 {
		return "unknown"
	}
	if float64(chineseCount)/float64(total) > 0.15 {
		return "zh"
	}
	return "en"
}

// ──────────────────────────────────────────────
// 构建 Prompt
// ──────────────────────────────────────────────

func buildPrompt(action, title, content string) (string, error) {
	// 防止 content 过长（DeepSeek 上下文限制），截取前 4000 个字符
	maxRunes := 4000
	runes := []rune(content)
	if len(runes) > maxRunes {
		content = string(runes[:maxRunes]) + "..."
	}

	switch action {
	case "summarize":
		lang := detectLanguage(content)
		replyLang := "中文"
		if lang == "en" {
			replyLang = "英文"
		}
		return fmt.Sprintf(
			"请用%s对以下文章进行简洁的摘要总结（150字以内），直接输出摘要内容，不要加前缀说明：\n\n标题：%s\n\n内容：%s",
			replyLang, title, content,
		), nil
	case "translate":
		lang := detectLanguage(content)
		if lang == "zh" {
			return fmt.Sprintf(
				"请将以下中文文章翻译成英文，保持原文格式，直接输出翻译结果，不要加前缀说明：\n\n标题：%s\n\n内容：%s",
				title, content,
			), nil
		}
		return fmt.Sprintf(
			"请将以下英文文章翻译成中文，保持原文格式，直接输出翻译结果，不要加前缀说明：\n\nTitle: %s\n\nContent: %s",
			title, content,
		), nil
	default:
		return "", fmt.Errorf("unsupported action: %s", action)
	}
}

// ──────────────────────────────────────────────
// Redis 缓存 Key
// ──────────────────────────────────────────────

func aiCacheKey(articleID, action string) string {
	return fmt.Sprintf("ai:%s:%s", articleID, action)
}

// ──────────────────────────────────────────────
// AIProcess — SSE 流式处理 Handler
//
// GET /api/articles/:id/ai?action=summarize|translate
// 响应格式：text/event-stream
//   data: <文字片段>\n\n
//   data: [DONE]\n\n
// ──────────────────────────────────────────────

func AIProcess(ctx *gin.Context) {
	id := ctx.Param("id")
	action := ctx.DefaultQuery("action", "summarize")
	if action != "summarize" && action != "translate" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "action must be 'summarize' or 'translate'"})
		return
	}

	// ── 1. 从数据库取文章 ──
	var article models.Article
	if err := global.Db.Where("id = ?", id).First(&article).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// ── 2. 设置 SSE 响应头 ──
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("X-Accel-Buffering", "no")

	flusher, ok := ctx.Writer.(http.Flusher)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "streaming not supported"})
		return
	}

	// ── 3. 检查 Redis 缓存 ──
	cKey := aiCacheKey(id, action)
	if global.RedisDB != nil {
		cached, err := global.RedisDB.Get(cKey).Result()
		if err == nil && utf8.ValidString(cached) && len(cached) > 0 {
			// 缓存命中：逐块模拟流式输出（每50字一帧）
			chunkSize := 50
			runes := []rune(cached)
			for i := 0; i < len(runes); i += chunkSize {
				end := i + chunkSize
				if end > len(runes) {
					end = len(runes)
				}
				chunk := string(runes[i:end])
				fmt.Fprintf(ctx.Writer, "data: %s\n\n", chunk)
				flusher.Flush()
			}
			fmt.Fprintf(ctx.Writer, "data: [DONE]\n\n")
			flusher.Flush()
			return
		} else if err != nil && err != redis.Nil {
			// Redis 故障不影响主流程，继续调用 AI
		}
	}

	// ── 4. 构建 Prompt ──
	prompt, err := buildPrompt(action, article.Title, article.Content)
	if err != nil {
		fmt.Fprintf(ctx.Writer, "event: error\ndata: %s\n\n", err.Error())
		flusher.Flush()
		return
	}

	// ── 5. 构造 DeepSeek 请求 ──
	dsReq := dsRequest{
		Model: config.Appconfig.DeepSeek.Model,
		Messages: []dsMessage{
			{Role: "user", Content: prompt},
		},
		Stream: true,
	}
	reqBody, err := json.Marshal(dsReq)
	if err != nil {
		fmt.Fprintf(ctx.Writer, "event: error\ndata: marshal error\n\n")
		flusher.Flush()
		return
	}

	httpReq, err := http.NewRequest("POST", config.Appconfig.DeepSeek.BaseUrl, bytes.NewReader(reqBody))
	if err != nil {
		fmt.Fprintf(ctx.Writer, "event: error\ndata: request创建失败\n\n")
		flusher.Flush()
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+config.Appconfig.DeepSeek.ApiKey)

	// ── 6. 发起请求，流式读取并转发给前端 ──
	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		fmt.Fprintf(ctx.Writer, "event: error\ndata: AI API 请求失败\n\n")
		flusher.Flush()
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Fprintf(ctx.Writer, "event: error\ndata: AI API 错误 %d: %s\n\n", resp.StatusCode, string(body))
		flusher.Flush()
		return
	}

	// ── 7. 逐行解析 SSE，转发给前端，同时拼接完整结果用于缓存 ──
	var fullResult strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "[DONE]" {
			break
		}

		var chunk dsChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		if len(chunk.Choices) == 0 {
			continue
		}
		delta := chunk.Choices[0].Delta.Content
		if delta == "" {
			continue
		}

		fullResult.WriteString(delta)

		// 转发给前端
		fmt.Fprintf(ctx.Writer, "data: %s\n\n", delta)
		flusher.Flush()
	}

	// 发送结束信号
	fmt.Fprintf(ctx.Writer, "data: [DONE]\n\n")
	flusher.Flush()

	// ── 8. 将完整结果写入 Redis（TTL 1小时）──
	if global.RedisDB != nil {
		result := fullResult.String()
		if len(result) > 0 {
			_ = global.RedisDB.Set(cKey, result, time.Hour).Err()
		}
	}
}
