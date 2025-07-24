package ptt

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"github.com/twtrubiks/ptt-spider-go/config"
	"github.com/twtrubiks/ptt-spider-go/constants"
)

// customTransport 是一個自訂的 http.RoundTripper，用於在請求中加入 User-Agent
type customTransport struct {
	transport http.RoundTripper
}

// RoundTrip 攔截請求，加入 User-Agent 標頭，然後繼續發送請求
func (t *customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", constants.DefaultUserAgent)
	transport := t.transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	return transport.RoundTrip(req)
}

// configureCookies 為客戶端配置 over18 cookie
func configureCookies(client *http.Client) error {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return fmt.Errorf("建立 cookie jar 失敗: %w", err)
	}

	// 設定 over18 cookie
	overEighteenURL, err := url.Parse(constants.Over18URL)
	if err != nil {
		return fmt.Errorf("解析 over18 URL 失敗: %w", err)
	}

	jar.SetCookies(overEighteenURL, []*http.Cookie{
		{Name: constants.Over18CookieName, Value: constants.Over18CookieValue},
	})

	client.Jar = jar
	return nil
}

// NewClient 建立一個新的 http 客戶端，並設定 over18 cookie
func NewClient() (*http.Client, error) {
	return newClientWithOptions(nil)
}

// NewClientWithConfig 建立一個新的 http 客戶端，使用指定的配置進行連線池優化
func NewClientWithConfig(cfg *config.Config) (*http.Client, error) {
	return newClientWithOptions(cfg)
}

// newClientWithOptions 統一的客戶端建立邏輯
func newClientWithOptions(cfg *config.Config) (*http.Client, error) {
	var transport http.RoundTripper

	if cfg != nil {
		// 建立優化的 HTTP Transport
		httpTransport := &http.Transport{
			MaxIdleConns:          cfg.Crawler.HTTP.MaxIdleConns,
			MaxIdleConnsPerHost:   cfg.Crawler.HTTP.MaxIdleConnsPerHost,
			IdleConnTimeout:       cfg.GetIdleConnTimeout(),
			TLSHandshakeTimeout:   cfg.GetTLSHandshakeTimeout(),
			ExpectContinueTimeout: cfg.GetExpectContinueTimeout(),
			DisableKeepAlives:     false, // 啟用 Keep-Alive
		}
		transport = &customTransport{transport: httpTransport}
	} else {
		transport = &customTransport{}
	}

	client := &http.Client{
		Transport: transport,
	}

	// 設定超時（僅在有配置時）
	if cfg != nil {
		client.Timeout = cfg.GetTimeoutDuration()
	}

	// 配置 cookies
	if err := configureCookies(client); err != nil {
		return nil, fmt.Errorf("配置 cookie 失敗: %w", err)
	}

	return client, nil
}
