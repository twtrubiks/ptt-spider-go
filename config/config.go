// Package config 提供配置檔案管理功能，支援 YAML 格式的配置檔案載入和解析.
// 配置系統支援自動降級機制，當配置檔案不存在或解析失敗時，會自動使用預設配置.
package config

import (
	"fmt"
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
// YAML 字串欄位在 Load 時會一次性解析為 time.Duration，避免重複解析。
type HTTPConfig struct {
	Timeout               string `yaml:"timeout"`               // HTTP 請求超時時間（YAML 字串）
	MaxIdleConns          int    `yaml:"maxIdleConns"`          // 最大空閒連線數
	MaxIdleConnsPerHost   int    `yaml:"maxIdleConnsPerHost"`   // 每個主機的最大空閒連線數
	IdleConnTimeout       string `yaml:"idleConnTimeout"`       // 空閒連線超時時間（YAML 字串）
	TLSHandshakeTimeout   string `yaml:"tlsHandshakeTimeout"`   // TLS 握手超時時間（YAML 字串）
	ExpectContinueTimeout string `yaml:"expectContinueTimeout"` // Expect: 100-continue 超時時間（YAML 字串）

	// 已解析的 duration 值，Load 後即可直接使用
	parsed                bool          `yaml:"-"`
	timeout               time.Duration `yaml:"-"`
	idleConnTimeout       time.Duration `yaml:"-"`
	tlsHandshakeTimeout   time.Duration `yaml:"-"`
	expectContinueTimeout time.Duration `yaml:"-"`
}

// parseDurationWithDefault 解析 duration 字串，失敗時記錄警告並回傳預設值。
func parseDurationWithDefault(value string, defaultVal time.Duration, name string) time.Duration {
	if d, err := time.ParseDuration(value); err == nil {
		return d
	}
	log.Printf("解析 %s 失敗，使用預設值 %v", name, defaultVal)
	return defaultVal
}

// parseHTTPDurations 一次性解析所有 HTTP duration 字串為 time.Duration。
func (h *HTTPConfig) parseHTTPDurations() {
	h.timeout = parseDurationWithDefault(h.Timeout, 30*time.Second, "HTTP 超時時間")
	h.idleConnTimeout = parseDurationWithDefault(h.IdleConnTimeout, 90*time.Second, "空閒連線超時時間")
	h.tlsHandshakeTimeout = parseDurationWithDefault(h.TLSHandshakeTimeout, 10*time.Second, "TLS 握手超時時間")
	h.expectContinueTimeout = parseDurationWithDefault(h.ExpectContinueTimeout, 1*time.Second, "Expect Continue 超時時間")
	h.parsed = true
}

// DefaultConfig 返回預設配置，包含適合一般使用場景的參數設定.
// 這些預設值經過測試，能在大多數環境下穩定運行.
func DefaultConfig() *Config {
	cfg := &Config{
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
	cfg.Crawler.HTTP.parseHTTPDurations()
	return cfg
}

// Load 載入配置檔案，支援自動降級機制.
// 當配置檔案不存在時，會自動使用預設配置並返回 nil error.
// 當讀取或解析失敗時，返回 error 讓呼叫方決定處理方式.
// 參數:
//   - configPath: 配置檔案的完整路徑
//
// 返回:
//   - *Config: 配置物件，檔案不存在時為預設配置，讀取/解析失敗時為 nil
//   - error: 讀取或解析失敗時返回對應錯誤
func Load(configPath string) (*Config, error) {
	// 如果配置檔案不存在，使用預設配置
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Printf("配置檔案 %s 不存在，使用預設配置", configPath)
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("讀取配置檔案失敗: %w", err)
	}

	config := DefaultConfig()
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("解析配置檔案失敗: %w", err)
	}

	// 一次性解析所有 duration 字串
	config.Crawler.HTTP.parseHTTPDurations()

	log.Printf("成功載入配置檔案: %s", configPath)
	return config, nil
}

// ensureParsed 確保 HTTP duration 已被解析。
// 支援直接建構 Config 結構體（不經過 Load）的使用場景。
func (c *Config) ensureParsed() {
	if !c.Crawler.HTTP.parsed {
		c.Crawler.HTTP.parseHTTPDurations()
	}
}

// GetTimeoutDuration 獲取已解析的 HTTP 請求超時時間.
func (c *Config) GetTimeoutDuration() time.Duration {
	c.ensureParsed()
	return c.Crawler.HTTP.timeout
}

// GetDelayRange 獲取延遲範圍，用於隨機延遲計算.
// 返回最小和最大延遲時間，用於避免對伺服器造成過大壓力.
func (c *Config) GetDelayRange() (time.Duration, time.Duration) {
	minDelay := time.Duration(c.Crawler.Delays.MinMs) * time.Millisecond
	maxDelay := time.Duration(c.Crawler.Delays.MaxMs) * time.Millisecond
	return minDelay, maxDelay
}

// GetIdleConnTimeout 獲取已解析的空閒連線超時時間.
func (c *Config) GetIdleConnTimeout() time.Duration {
	c.ensureParsed()
	return c.Crawler.HTTP.idleConnTimeout
}

// GetTLSHandshakeTimeout 獲取已解析的 TLS 握手超時時間.
func (c *Config) GetTLSHandshakeTimeout() time.Duration {
	c.ensureParsed()
	return c.Crawler.HTTP.tlsHandshakeTimeout
}

// GetExpectContinueTimeout 獲取已解析的 Expect: 100-continue 超時時間.
func (c *Config) GetExpectContinueTimeout() time.Duration {
	c.ensureParsed()
	return c.Crawler.HTTP.expectContinueTimeout
}
