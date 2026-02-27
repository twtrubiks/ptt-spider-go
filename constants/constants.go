// Package constants 定義 PTT 爬蟲使用的常數值，
// 包括 URL、HTTP 標頭、檔案權限和預設配置。
package constants

const (
	// PTT URLs
	PttBaseURL = "https://www.ptt.cc"
	Over18URL  = "https://www.ptt.cc/ask/over18"

	// HTTP Headers
	DefaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"

	// File permissions
	DirPermission  = 0755
	FilePermission = 0644

	// Default values
	DefaultBoard    = "beauty"
	DefaultPages    = 3
	DefaultPushRate = 10

	// Over18 cookie settings
	Over18CookieName  = "over18"
	Over18CookieValue = "1"

	// HTTP 429 retry settings
	RetryMaxAttempts    = 3     // 最多重試次數
	RetryInitialDelayMs = 1000  // 初始退避延遲（毫秒）
	RetryMaxDelayMs     = 30000 // 最大退避延遲（毫秒）
	RetryBackoffFactor  = 2     // 指數退避倍數
)
