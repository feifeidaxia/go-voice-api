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

func TTSHandler(c *gin.Context) {
	var req TTSRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Text == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if req.Voice == "" {
		req.Voice = "nova" // 默认语音
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "API key not set"})
		return
	}

	body := map[string]interface{}{
		"model":  "gpt-4o-mini-tts", // 流式支持模型
		"input":  req.Text,
		"voice":  req.Voice,
		"stream": req.Stream,
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

	resp, err := utils.HTTPClient.Do(httpReq)
	if err != nil || resp.StatusCode != 200 {
		log.Println("Error calling OpenAI TTS API:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to call OpenAI TTS API"})
		return
	}
	defer resp.Body.Close()
	if req.Stream {
		c.Header("Content-Type", "audio/mpeg")
		c.Status(http.StatusOK)

		buf := make([]byte, 1024)
		for {
			n, err := resp.Body.Read(buf)
			if n > 0 {
				if _, werr := c.Writer.Write(buf[:n]); werr != nil {
					log.Println("Error writing chunk:", werr)
					break
				}
				c.Writer.Flush()
			}
			if err != nil {
				if err != io.EOF {
					log.Println("Read error:", err)
				}
				break
			}
		}
	} else {
		// 非流式
		audioData, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println("Error reading audio response:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read audio response"})
			return
		}
		c.Data(http.StatusOK, "audio/mpeg", audioData)
	}
}
