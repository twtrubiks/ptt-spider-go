package interfaces

import (
	"context"
	"io"
	"net/http"

	"github.com/twtrubiks/ptt-spider-go/types"
)

// HTTPClient 定義 HTTP 客戶端介面，便於測試和替換實現
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Parser 定義解析器介面，用於解析 PTT 頁面內容
type Parser interface {
	// ParseArticles 解析文章列表頁面
	ParseArticles(body io.Reader) ([]types.ArticleInfo, error)
	// ParseArticleContent 解析文章內容頁面，返回標題和圖片 URLs
	ParseArticleContent(body io.Reader) (title string, imageURLs []string, err error)
	// GetMaxPage 獲取看板的最大頁數
	GetMaxPage(ctx context.Context, client HTTPClient, board string) (int, error)
}

// MarkdownGenerator 定義 Markdown 生成器介面
type MarkdownGenerator interface {
	// Generate 根據提供的資訊生成 Markdown 檔案
	Generate(info types.MarkdownInfo) error
}

// FileDownloader 定義檔案下載器介面
type FileDownloader interface {
	// Download 下載檔案到指定路徑
	Download(ctx context.Context, url, savePath string) error
}

// ConfigLoader 定義配置載入器介面
type ConfigLoader interface {
	// Load 從指定路徑載入配置
	Load(path string) error
	// Get 獲取配置值
	Get(key string) interface{}
	// Set 設定配置值
	Set(key string, value interface{})
}

// ArticleProducer 定義文章生產者介面
type ArticleProducer interface {
	// ProduceFromBoard 從看板產生文章列表
	ProduceFromBoard(ctx context.Context, board string, pages int, pushRate int) (<-chan types.ArticleInfo, error)
	// ProduceFromFile 從檔案產生文章列表
	ProduceFromFile(ctx context.Context, filePath string) (<-chan types.ArticleInfo, error)
}

// ContentProcessor 定義內容處理器介面
type ContentProcessor interface {
	// Process 處理文章內容，返回下載任務和 Markdown 任務
	Process(ctx context.Context, article types.ArticleInfo) ([]types.DownloadTask, *types.MarkdownInfo, error)
}

// WorkerPool 定義工人池介面
type WorkerPool interface {
	// Start 啟動工人池
	Start(ctx context.Context)
	// Stop 停止工人池
	Stop()
	// Submit 提交任務到工人池
	Submit(task interface{}) error
	// Wait 等待所有任務完成
	Wait()
}

// Logger 定義日誌記錄器介面
type Logger interface {
	// Debug 記錄除錯訊息
	Debug(msg string, fields ...interface{})
	// Info 記錄資訊訊息
	Info(msg string, fields ...interface{})
	// Warn 記錄警告訊息
	Warn(msg string, fields ...interface{})
	// Error 記錄錯誤訊息
	Error(msg string, fields ...interface{})
	// Fatal 記錄致命錯誤並退出
	Fatal(msg string, fields ...interface{})
}

// Crawler 定義爬蟲介面
type Crawler interface {
	// Run 啟動爬蟲
	Run(ctx context.Context) error
	// SetConfig 設定爬蟲配置
	SetConfig(config interface{})
	// GetStats 獲取爬蟲統計資訊
	GetStats() map[string]interface{}
}

// Validator 定義驗證器介面
type Validator interface {
	// ValidateURL 驗證 URL 格式
	ValidateURL(url string) error
	// ValidateBoard 驗證看板名稱
	ValidateBoard(board string) error
	// ValidateConfig 驗證配置
	ValidateConfig(config interface{}) error
}

// CacheManager 定義快取管理器介面
type CacheManager interface {
	// Get 從快取獲取值
	Get(key string) (interface{}, bool)
	// Set 設定快取值
	Set(key string, value interface{})
	// Delete 刪除快取值
	Delete(key string)
	// Clear 清空快取
	Clear()
}

// RateLimiter 定義速率限制器介面
type RateLimiter interface {
	// Allow 檢查是否允許執行操作
	Allow() bool
	// Wait 等待直到允許執行操作
	Wait(ctx context.Context) error
}

// MetricsCollector 定義指標收集器介面
type MetricsCollector interface {
	// RecordRequest 記錄請求
	RecordRequest(method, endpoint string, duration int64)
	// RecordError 記錄錯誤
	RecordError(errorType string)
	// GetMetrics 獲取指標
	GetMetrics() map[string]interface{}
}
