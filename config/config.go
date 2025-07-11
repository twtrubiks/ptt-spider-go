// Package config 提供配置檔案管理功能，支援 YAML 格式的配置檔案載入和解析.
// 配置系統支援自動降級機制，當配置檔案不存在或解析失敗時，會自動使用預設配置.
package config

import (
	"log"
	"os"
	"time"

	yaml "gopkg.in/yaml.v3"
)

// Config 整體配置結構，包含爬蟲的所有配置項目.
type Config struct {
	Crawler CrawlerConfig `yaml:"crawler"`
}

// CrawlerConfig 爬蟲核心配置，控制並行度、通道大小和網路設定.
type CrawlerConfig struct {
	Workers     int           `yaml:"workers"`     // 並行下載工作者數量
	ParserCount int           `yaml:"parserCount"` // 內容解析器數量
	Channels    ChannelConfig `yaml:"channels"`    // 通道緩衝區配置
	Delays      DelayConfig   `yaml:"delays"`      // 延遲設定
	HTTP        HTTPConfig    `yaml:"http"`        // HTTP 客戶端配置
}

// ChannelConfig 通道緩衝區配置，用於控制 Goroutine 間的通訊容量.
type ChannelConfig struct {
	ArticleInfo  int `yaml:"articleInfo"`  // 文章資訊通道緩衝區大小
	DownloadTask int `yaml:"downloadTask"` // 下載任務通道緩衝區大小
	MarkdownTask int `yaml:"markdownTask"` // Markdown 任務通道緩衝區大小
}

// DelayConfig 延遲配置，用於控制請求間隔以避免被伺服器封鎖.
type DelayConfig struct {
	MinMs int `yaml:"minMs"` // 最小延遲毫秒數
	MaxMs int `yaml:"maxMs"` // 最大延遲毫秒數
}

// HTTPConfig HTTP 客戶端配置，控制連線超時和連線池設定.
type HTTPConfig struct {
	Timeout               string `yaml:"timeout"`               // HTTP 請求超時時間
	MaxIdleConns          int    `yaml:"maxIdleConns"`          // 最大空閒連線數
	MaxIdleConnsPerHost   int    `yaml:"maxIdleConnsPerHost"`   // 每個主機的最大空閒連線數
	IdleConnTimeout       string `yaml:"idleConnTimeout"`       // 空閒連線超時時間
	TLSHandshakeTimeout   string `yaml:"tlsHandshakeTimeout"`   // TLS 握手超時時間
	ExpectContinueTimeout string `yaml:"expectContinueTimeout"` // Expect: 100-continue 超時時間
}

// DefaultConfig 返回預設配置，包含適合一般使用場景的參數設定.
// 這些預設值經過測試，能在大多數環境下穩定運行.
func DefaultConfig() *Config {
	return &Config{
		Crawler: CrawlerConfig{
			Workers:     10,
			ParserCount: 10,
			Channels: ChannelConfig{
				ArticleInfo:  100,
				DownloadTask: 200,
				MarkdownTask: 100,
			},
			Delays: DelayConfig{
				MinMs: 500,
				MaxMs: 2000,
			},
			HTTP: HTTPConfig{
				Timeout:               "30s",
				MaxIdleConns:          100,
				MaxIdleConnsPerHost:   20,
				IdleConnTimeout:       "90s",
				TLSHandshakeTimeout:   "10s",
				ExpectContinueTimeout: "1s",
			},
		},
	}
}

// Load 載入配置檔案，支援自動降級機制.
// 當配置檔案不存在、讀取失敗或解析失敗時，會自動使用預設配置.
// 參數:
//   - configPath: 配置檔案的完整路徑
//
// 返回:
//   - *Config: 配置物件，不會為 nil
//   - error: 總是返回 nil，錯誤會被自動處理
func Load(configPath string) (*Config, error) {
	// 如果配置檔案不存在，使用預設配置
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Printf("配置檔案 %s 不存在，使用預設配置", configPath)
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("讀取配置檔案失敗，使用預設配置: %v", err)
		return DefaultConfig(), nil
	}

	config := DefaultConfig()
	if err := yaml.Unmarshal(data, config); err != nil {
		log.Printf("解析配置檔案失敗，使用預設配置: %v", err)
		return DefaultConfig(), nil
	}

	log.Printf("成功載入配置檔案: %s", configPath)
	return config, nil
}

// GetTimeoutDuration 獲取 HTTP 請求超時時間.
// 解析配置中的超時字符串，如果解析失敗則返回預設的 30 秒.
func (c *Config) GetTimeoutDuration() time.Duration {
	duration, err := time.ParseDuration(c.Crawler.HTTP.Timeout)
	if err != nil {
		log.Printf("解析超時時間失敗，使用預設值 30s: %v", err)
		return 30 * time.Second
	}
	return duration
}

// GetDelayRange 獲取延遲範圍，用於隨機延遲計算.
// 返回最小和最大延遲時間，用於避免對伺服器造成過大壓力.
func (c *Config) GetDelayRange() (time.Duration, time.Duration) {
	minDelay := time.Duration(c.Crawler.Delays.MinMs) * time.Millisecond
	maxDelay := time.Duration(c.Crawler.Delays.MaxMs) * time.Millisecond
	return minDelay, maxDelay
}

// GetIdleConnTimeout 獲取空閒連線超時時間.
// 解析配置中的空閒連線超時字符串，如果解析失敗則返回預設的 90 秒.
func (c *Config) GetIdleConnTimeout() time.Duration {
	duration, err := time.ParseDuration(c.Crawler.HTTP.IdleConnTimeout)
	if err != nil {
		log.Printf("解析空閒連線超時時間失敗，使用預設值 90s: %v", err)
		return 90 * time.Second
	}
	return duration
}

// GetTLSHandshakeTimeout 獲取 TLS 握手超時時間.
// 解析配置中的 TLS 握手超時字符串，如果解析失敗則返回預設的 10 秒.
func (c *Config) GetTLSHandshakeTimeout() time.Duration {
	duration, err := time.ParseDuration(c.Crawler.HTTP.TLSHandshakeTimeout)
	if err != nil {
		log.Printf("解析 TLS 握手超時時間失敗，使用預設值 10s: %v", err)
		return 10 * time.Second
	}
	return duration
}

// GetExpectContinueTimeout 獲取 Expect: 100-continue 超時時間.
// 解析配置中的 Expect 超時字符串，如果解析失敗則返回預設的 1 秒.
func (c *Config) GetExpectContinueTimeout() time.Duration {
	duration, err := time.ParseDuration(c.Crawler.HTTP.ExpectContinueTimeout)
	if err != nil {
		log.Printf("解析 Expect Continue 超時時間失敗，使用預設值 1s: %v", err)
		return 1 * time.Second
	}
	return duration
}
