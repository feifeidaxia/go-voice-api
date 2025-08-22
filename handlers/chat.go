package handlers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"go-voice-api/utils"

	"github.com/gin-gonic/gin"
)

// ChatRequest 接收前端传来的聊天请求（支持上下文 + 可选参数）
type ChatRequest struct {
	Messages    []OpenAIChatMessage `json:"messages"`              // 前端必须传 messages
	Model       *string             `json:"model,omitempty"`       // 可选，前端可覆盖模型
	MaxTokens   *int                `json:"max_tokens,omitempty"`  // 可选
	Temperature *float32            `json:"temperature,omitempty"` // 可选
}

// OpenAIChatMessage 表示单条聊天消息
type OpenAIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIChatRequest 发送给 OpenAI 的请求
type OpenAIChatRequest struct {
	Model       string              `json:"model"`
	Messages    []OpenAIChatMessage `json:"messages"`
	MaxTokens   int                 `json:"max_tokens,omitempty"`
	Temperature float32             `json:"temperature,omitempty"`
}

// OpenAIChatResponse 表示 OpenAI 的返回
type OpenAIChatResponse struct {
	Choices []struct {
		Message OpenAIChatMessage `json:"message"`
	} `json:"choices"`
}

// ChatHandler 处理聊天请求
func ChatHandler(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// 手动检查 messages
	if req.Messages == nil || len(req.Messages) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "messages cannot be empty"})
		return
	}

	// 默认值
	maxTokens := 1000
	temperature := float32(0.7)
	model := "gpt-3.5-turbo" // 默认模型

	// 如果前端传了，就覆盖默认值
	if req.MaxTokens != nil {
		maxTokens = *req.MaxTokens
	}
	if req.Temperature != nil {
		temperature = *req.Temperature
	}
	if req.Model != nil && *req.Model != "" {
		model = *req.Model
	}

	openaiReq := OpenAIChatRequest{
		Model:       model,
		Messages:    req.Messages,
		MaxTokens:   maxTokens,
		Temperature: temperature,
	}

	payload, err := json.Marshal(openaiReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal request"})
		return
	}

	openaiKey := os.Getenv("OPENAI_API_KEY")
	if openaiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "OpenAI API key not set"})
		return
	}

	reqOpenAI, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(payload))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}
	reqOpenAI.Header.Set("Authorization", "Bearer "+openaiKey)
	reqOpenAI.Header.Set("Content-Type", "application/json")

	resp, err := utils.HTTPClient.Do(reqOpenAI)
	if err != nil {
		log.Println("Request error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "OpenAI request failed"})
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
		return
	}

	if resp.StatusCode != 200 {
		log.Println("OpenAI returned error:", string(body))
		c.JSON(resp.StatusCode, gin.H{"error": "OpenAI API returned error", "detail": string(body)})
		return
	}

	var aiResp OpenAIChatResponse
	if err := json.Unmarshal(body, &aiResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse OpenAI response"})
		return
	}

	if len(aiResp.Choices) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No response from OpenAI"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"reply": aiResp.Choices[0].Message.Content,
	})
}
