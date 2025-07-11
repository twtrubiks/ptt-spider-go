package ptt

import (
	"net/http"
	"net/http/cookiejar"

	"github.com/twtrubiks/ptt-spider-go/config"
)

// customTransport 是一個自訂的 http.RoundTripper，用於在請求中加入 User-Agent
type customTransport struct {
	transport http.RoundTripper
}

// RoundTrip 攔截請求，加入 User-Agent 標頭，然後繼續發送請求
func (t *customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	transport := t.transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	return transport.RoundTrip(req)
}

// NewClient 建立一個新的 http 客戶端，並設定 over18 cookie
func NewClient() (*http.Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Jar:       jar,
		Transport: &customTransport{},
	}

	// 建立一個請求來設定 cookie
	req, err := http.NewRequest("GET", "https://www.ptt.cc/ask/over18?from=%2Fbbs%2FBeauty%2Findex.html", nil)
	if err != nil {
		return nil, err
	}
	// 注意：這裡的 AddCookie 是將 cookie 加到單次請求中。
	// 因為 client 的 Jar 已經設定，後續對 ptt.cc 的請求會自動帶上 over18=1 的 cookie。
	req.AddCookie(&http.Cookie{Name: "over18", Value: "1"})

	// 發送請求以確保 cookie 被設定
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	// 關閉回應主體以重用連接
	defer resp.Body.Close()

	return client, nil
}

// NewClientWithConfig 建立一個新的 http 客戶端，使用指定的配置進行連線池優化
func NewClientWithConfig(cfg *config.Config) (*http.Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	// 建立優化的 HTTP Transport
	transport := &http.Transport{
		MaxIdleConns:          cfg.Crawler.HTTP.MaxIdleConns,
		MaxIdleConnsPerHost:   cfg.Crawler.HTTP.MaxIdleConnsPerHost,
		IdleConnTimeout:       cfg.GetIdleConnTimeout(),
		TLSHandshakeTimeout:   cfg.GetTLSHandshakeTimeout(),
		ExpectContinueTimeout: cfg.GetExpectContinueTimeout(),
		DisableKeepAlives:     false, // 啟用 Keep-Alive
	}

	client := &http.Client{
		Jar:       jar,
		Timeout:   cfg.GetTimeoutDuration(),
		Transport: &customTransport{transport: transport},
	}

	// 建立一個請求來設定 cookie
	req, err := http.NewRequest("GET", "https://www.ptt.cc/ask/over18?from=%2Fbbs%2FBeauty%2Findex.html", nil)
	if err != nil {
		return nil, err
	}
	// 注意：這裡的 AddCookie 是將 cookie 加到單次請求中。
	// 因為 client 的 Jar 已經設定，後續對 ptt.cc 的請求會自動帶上 over18=1 的 cookie。
	req.AddCookie(&http.Cookie{Name: "over18", Value: "1"})

	// 發送請求以確保 cookie 被設定
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	// 關閉回應主體以重用連接
	defer resp.Body.Close()

	return client, nil
}
