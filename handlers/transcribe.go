package handlers

import (
	"bytes"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"

	"go-voice-api/utils"

	"github.com/gin-gonic/gin"
)

// Response 表示成功响应结构
type Response struct {
	Text string `json:"text" example:"hello world"`
}

// ErrorResponse 表示失败响应结构
type ErrorResponse struct {
	Error string `json:"error" example:"something went wrong"`
}

// @Summary Transcribes an audio file
// @Description This API transcribes audio to text
// @Tags transcription
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Audio file to be transcribed"
// @Success 200 {object} Response
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/transcribe [post]
func TranscribeHandler(c *gin.Context) {
	// 获取上传的音频文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Audio file is required"})
		return
	}
	defer file.Close()

	// 读取文件内容
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	formFile, err := writer.CreateFormFile("file", header.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create form file"})
		return
	}
	io.Copy(formFile, file)

	// 添加 Whisper 所需的参数
	writer.WriteField("model", "whisper-1")
	writer.Close()

	// 读取 OpenAI API Key
	openaiKey := os.Getenv("OPENAI_API_KEY")
	if openaiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "API key not set"})
		return
	}

	// 发送 POST 请求到 OpenAI
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/audio/transcriptions", &buf)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}
	req.Header.Set("Authorization", "Bearer "+openaiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := utils.HTTPClient.Do(req)
	if err != nil {
		log.Println("Request error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to call OpenAI API"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Println("OpenAI error:", string(body))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "OpenAI API error"})
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
		return
	}

	// 返回 OpenAI 的结果
	c.Data(http.StatusOK, "application/json", body)
}
