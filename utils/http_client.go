package utils

import (
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"
)

var HTTPClient *http.Client

func InitHTTPClient() {
	useProxy := os.Getenv("USE_PROXY")
	var transport *http.Transport

	if useProxy == "true" {
		proxyURL, _ := url.Parse("http://127.0.0.1:6987") // 本地 HTTP 代理
		log.Println("✅ USE_PROXY=true，启用本地代理：", proxyURL)

		transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout: 10 * time.Second,
		}
	} else {
		log.Println("❌ USE_PROXY=false，跳过代理，直接连接 OpenAI")
		transport = &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout: 10 * time.Second,
		}
	}

	HTTPClient = &http.Client{
		Timeout:   60 * time.Second,
		Transport: transport,
	}
}
