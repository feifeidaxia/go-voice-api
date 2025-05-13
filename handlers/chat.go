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

// ChatRequest 接收前端传来的聊天请求，支持完整历史消息
type ChatRequest struct {
	Messages []OpenAIChatMessage `json:"messages"` // 聊天记录数组（包括历史消息）
}

// OpenAIChatMessage 表示单条聊天消息
type OpenAIChatMessage struct {
	Role    string `json:"role"`    // "user"、"assistant" 或 "system"
	Content string `json:"content"` // 消息内容
}

// OpenAIChatRequest 是发送给 OpenAI 的请求结构
type OpenAIChatRequest struct {
	Model    string              `json:"model"`    // 模型名称，例如 "gpt-3.5-turbo"
	Messages []OpenAIChatMessage `json:"messages"` // 消息数组
}

// OpenAIChatResponse 表示 OpenAI 返回的响应结构
type OpenAIChatResponse struct {
	Choices []struct {
		Message OpenAIChatMessage `json:"message"`
	} `json:"choices"`
}

// ChatHandler 处理聊天请求
// @Summary 聊天接口
// @Description 使用 OpenAI GPT 与用户对话（支持上下文消息）
// @Tags Chat
// @Accept json
// @Produce json
// @Param request body ChatRequest true "聊天消息数组（包括历史消息）"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/chat [post]
func ChatHandler(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	openaiReq := OpenAIChatRequest{
		Model:    "gpt-3.5-turbo",
		Messages: req.Messages,
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

	// 使用封装的 HTTP 客户端（包含代理配置）
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
