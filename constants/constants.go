// Package constants 定義 PTT 爬蟲使用的常數值，
// 包括 URL、HTTP 標頭、檔案權限和預設配置。
package constants

const (
	// PttBaseURL 是 PTT 網站的基底 URL
	PttBaseURL = "https://www.ptt.cc"
	// Over18URL 是十八禁確認頁的 URL
	Over18URL = "https://www.ptt.cc/ask/over18"

	// DefaultUserAgent 是 HTTP 請求使用的 User-Agent 標頭
	DefaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"

	// DirPermission 是建立目錄使用的權限
	DirPermission = 0755
	// FilePermission 是建立檔案使用的權限
	FilePermission = 0644

	// DefaultBoard 是預設的看板名稱
	DefaultBoard = "beauty"
	// DefaultPages 是預設的爬取頁數
	DefaultPages = 3
	// DefaultPushRate 是預設的推文數門檻
	DefaultPushRate = 10

	// Over18CookieName 是十八禁確認 cookie 的名稱
	Over18CookieName = "over18"
	// Over18CookieValue 是十八禁確認 cookie 的值
	Over18CookieValue = "1"

	// RetryMaxAttempts 是 HTTP 429 的最多重試次數
	RetryMaxAttempts = 3
	// RetryInitialDelayMs 是重試的初始退避延遲（毫秒）
	RetryInitialDelayMs = 1000
	// RetryMaxDelayMs 是重試的最大退避延遲（毫秒）
	RetryMaxDelayMs = 30000
	// RetryBackoffFactor 是重試的指數退避倍數
	RetryBackoffFactor = 2

	// MaxImageSizeBytes 單張圖片下載大小上限（50 MB），
	// 圖片連結來自文章內容（外部可控），防止超大回應寫爆磁碟
	MaxImageSizeBytes int64 = 50 * 1024 * 1024
)
