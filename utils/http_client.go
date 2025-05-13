// utils/http_client.go
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
	proxyURL, _ := url.Parse("http://127.0.0.1:6987") // HTTP 代理地址
	log.Println("USE_PROXY =", os.Getenv("USE_PROXY"))

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	log.Println("🚀 HTTP 代理已配置: ", proxyURL)

	HTTPClient = &http.Client{
		Timeout:   60 * time.Second,
		Transport: transport,
	}
}
