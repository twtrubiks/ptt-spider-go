// Package types 定義 PTT Spider 的核心資料結構與進度事件類型。
package types

// EventType 定義進度事件的類型
type EventType int

// 進度事件類型常數
const (
	EventPageParsed    EventType = iota // 列表頁解析完成
	EventArticleParsed                  // 文章解析完成
	EventDownloadStart                  // 開始下載圖片
	EventDownloadDone                   // 圖片下載完成
	EventDownloadFail                   // 圖片下載失敗
	EventCrawlerDone                    // 爬蟲全部完成
)

// ProgressEvent 表示爬蟲執行過程中的進度事件
type ProgressEvent struct {
	Type         EventType // 事件類型
	WorkerID     int       // Worker 編號（僅下載事件使用）
	Message      string    // 事件描述訊息
	ArticleTitle string    // 文章標題（僅文章相關事件使用）
	ImageCount   int       // 圖片數量（僅 EventArticleParsed 使用）
	CurrentPage  int       // 當前頁數（僅 EventPageParsed 使用）
	TotalPages   int       // 總頁數（僅 EventPageParsed 使用）
}
