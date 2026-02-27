package interfaces

import (
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
	// ParseMaxPage 從看板首頁 HTML 解析最大頁數
	ParseMaxPage(body io.Reader) (int, error)
}

// MarkdownGenerator 定義 Markdown 生成器介面
type MarkdownGenerator interface {
	// Generate 根據提供的資訊生成 Markdown 檔案
	Generate(info types.MarkdownInfo) error
}
