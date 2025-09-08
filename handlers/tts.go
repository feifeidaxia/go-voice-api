package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"go-voice-api/utils"

	"github.com/gin-gonic/gin"
)

type TTSRequest struct {
	Text   string `json:"text"`
	Voice  string `json:"voice"`  // 用户选择的语音风格
	Stream bool   `json:"stream"` // 是否流式返回
}

// @Summary Converts text to speech
// @Description This API converts provided text into speech audio
// @Tags tts
// @Accept json
// @Produce audio/mpeg
// @Param text body TTSRequest true "Text to be converted to speech"
// @Success 200 {file} AudioFile "Audio file containing the speech"
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/tts [post]
func TTSHandler(c *gin.Context) {
	var req TTSRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Text == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// 默认 voice
	if req.Voice == "" {
		req.Voice = "nova"
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "API key not set"})
		return
	}

	// 调用 OpenAI 的 TTS 接口
	body := map[string]interface{}{
		"model":  "tts-1",
		"input":  req.Text,
		"voice":  req.Voice,
		"stream": req.Stream, // 只要传 true，就会流式
	}
	jsonBody, _ := json.Marshal(body)

	reqUrl := "https://api.openai.com/v1/audio/speech"
	httpReq, err := http.NewRequest("POST", reqUrl, bytes.NewReader(jsonBody))
	if err != nil {
		log.Println("Error creating request:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	// 使用封装的 HTTP 客户端（包含代理配置）
	resp, err := utils.HTTPClient.Do(httpReq)
	if err != nil || resp.StatusCode != 200 {
		log.Println("Error calling OpenAI TTS API:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to call OpenAI TTS API"})
		return
	}
	defer resp.Body.Close()

	if req.Stream {
		// 流式传输
		c.Header("Content-Type", "audio/mpeg")
		c.Status(http.StatusOK)

		// 将 OpenAI 响应体直接 copy 给前端
		if _, err := io.Copy(c.Writer, resp.Body); err != nil {
			log.Println("Error streaming audio:", err)
		}
	} else {
		// 非流式，一次性读完
		audioData, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println("Error reading audio response:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read audio response"})
			return
		}
		c.Data(http.StatusOK, "audio/mpeg", audioData)
	}
}
