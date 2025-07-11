package types

// ArticleInfo 用於儲存從看板列表頁解析出的基本文章資訊.
type ArticleInfo struct {
	Title    string // 文章標題，可能為空（檔案模式時）
	URL      string // 文章完整 URL
	Author   string // 作者帳號
	PushRate int    // 推文數（正數為推，負數為噓）
}

// DownloadTask 用於儲存單一圖片的下載任務資訊.
type DownloadTask struct {
	ImageURL string // 圖片的完整 URL
	SavePath string // 圖片應儲存的完整本地路徑 (含檔名)
}

// MarkdownInfo 用於儲存產生 Markdown 檔案所需的資訊.
type MarkdownInfo struct {
	Title      string   // 文章標題
	ArticleURL string   // 原始文章 URL
	PushCount  int      // 推文數
	ImageURLs  []string // 所有圖片 URL 列表
	SaveDir    string   // 儲存 Markdown 和圖片的目錄
}
