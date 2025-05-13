package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	_ "go-voice-api/docs" // Swagger 自动生成文档用

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"go-voice-api/handlers"
	"go-voice-api/utils"
)

// @title           Go Voice API
// @version         1.0
// @description     基于 OpenAI 的语音助手 API 接口
// @host            localhost:8080
// @BasePath        /
func main() {
	// 初始化 HTTP 客户端（含代理判断）
	utils.InitHTTPClient()

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system env")
	}

	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Server running"})
	})

	// 注册 Swagger 文档路由
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := router.Group("/api")
	{
		api.POST("/transcribe", handlers.TranscribeHandler)
		api.POST("/chat", handlers.ChatHandler)
		api.POST("/tts", handlers.TTSHandler)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("🚀 服务器启动成功，监听端口: %s", port)
	router.Run(":" + port)
}
