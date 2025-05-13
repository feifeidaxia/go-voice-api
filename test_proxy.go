package main

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		fmt.Println("⚠️  无法加载 .env 文件，使用系统环境变量")
	}

	// 获取 OpenAI API Key
	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		fmt.Println("❌ OPENAI_API_KEY 未设置")
		return
	}

	// 设置本地代理（http://127.0.0.1:6987）
	proxyURL, _ := url.Parse("http://127.0.0.1:6987")
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	client := &http.Client{
		Timeout:   60 * time.Second,
		Transport: transport,
	}

	// 创建请求
	req, _ := http.NewRequest("GET", "https://api.openai.com/v1/models", nil)
	req.Header.Set("Authorization", "Bearer "+key)

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("❌ 请求失败：", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("✅ 请求成功，状态码：", resp.StatusCode)
}
