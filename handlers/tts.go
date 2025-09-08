package handlers

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

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
		req.Voice = "nova"
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "API key not set"})
		return
	}

	body := map[string]interface{}{
		"model":  "gpt-4o-mini-tts",
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
		c.Header("Cache-Control", "no-cache")
		c.Writer.WriteHeader(http.StatusOK)

		// 使用 bufio.Scanner 逐行解析 SSE
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "data: ") {
				payload := strings.TrimPrefix(line, "data: ")
				if payload == "[DONE]" {
					break
				}

				var chunk struct {
					Audio string `json:"audio"` // SSE 返回的 base64
				}
				if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
					log.Println("JSON parse error:", err)
					continue
				}

				audioBytes, err := base64.StdEncoding.DecodeString(chunk.Audio)
				if err != nil {
					log.Println("Base64 decode error:", err)
					continue
				}

				if _, werr := c.Writer.Write(audioBytes); werr != nil {
					log.Println("Write error:", werr)
					break
				}
				if flusher, ok := c.Writer.(http.Flusher); ok {
					flusher.Flush()
				}
			}
		}

		if err := scanner.Err(); err != nil {
			log.Println("Scanner error:", err)
		}
	} else {
		// 非流式直接返回
		audioData, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println("Error reading audio response:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read audio response"})
			return
		}
		c.Data(http.StatusOK, "audio/mpeg", audioData)
	}
}
