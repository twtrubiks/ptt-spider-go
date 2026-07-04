package crawler

import (
	"context"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/twtrubiks/ptt-spider-go/config"
	"github.com/twtrubiks/ptt-spider-go/mocks"
	"github.com/twtrubiks/ptt-spider-go/types"
)

func TestUniqueDirName(t *testing.T) {
	c := &Crawler{}

	// 第一篇文章取得原始目錄名
	if got := c.uniqueDirName("標題_10", "http://example.com/a1"); got != "標題_10" {
		t.Errorf("first article dir = %q, want %q", got, "標題_10")
	}

	// 同一篇文章（相同 URL）重複處理應回傳相同目錄
	if got := c.uniqueDirName("標題_10", "http://example.com/a1"); got != "標題_10" {
		t.Errorf("same article dir = %q, want %q", got, "標題_10")
	}

	// 不同文章撞名時加序號後綴
	if got := c.uniqueDirName("標題_10", "http://example.com/a2"); got != "標題_10_2" {
		t.Errorf("second article dir = %q, want %q", got, "標題_10_2")
	}
	if got := c.uniqueDirName("標題_10", "http://example.com/a3"); got != "標題_10_3" {
		t.Errorf("third article dir = %q, want %q", got, "標題_10_3")
	}
}

// newCollisionTestCrawler 建立以 mock 回傳指定圖片 URL 的 Crawler
func newCollisionTestCrawler(imgURLs []string) *Crawler {
	mockClient := &mocks.MockHTTPClient{
		DoFunc: func(_ *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader("<html></html>")),
			}, nil
		},
	}
	mockParser := &mocks.MockParser{
		ParseArticleContentFunc: func(_ io.Reader) (string, []string, error) {
			return "Same Title", imgURLs, nil
		},
	}
	cfg := &config.Config{
		Crawler: config.CrawlerConfig{
			Delays: config.DelayConfig{MinMs: 0, MaxMs: 0},
		},
	}
	return NewCrawlerWithDependencies(mockClient, mockParser, mocks.NewMockMarkdownGenerator(), "test", 1, 0, "", cfg)
}

// TestProcessArticle_SameBaseNameGetsSuffix 驗證不同 URL 推導出相同檔名時，
// 下載任務的儲存路徑會加上序號後綴而不會互相覆蓋。
func TestProcessArticle_SameBaseNameGetsSuffix(t *testing.T) {
	c := newCollisionTestCrawler([]string{
		"http://host1.com/a.jpg",
		"http://host2.com/a.jpg",
	})

	downloadChan := make(chan types.DownloadTask, 10)
	markdownChan := make(chan types.MarkdownInfo, 10)
	article := types.ArticleInfo{Title: "Same Title", URL: "http://example.com/article", PushRate: 10}
	c.processArticle(context.Background(), article, downloadChan, markdownChan)

	close(downloadChan)
	var paths []string
	for task := range downloadChan {
		paths = append(paths, task.SavePath)
	}

	if len(paths) != 2 {
		t.Fatalf("expected 2 download tasks, got %d", len(paths))
	}

	wantDir := filepath.Join("test", "Same Title_10")
	if paths[0] != filepath.Join(wantDir, "a.jpg") {
		t.Errorf("first save path = %q, want %q", paths[0], filepath.Join(wantDir, "a.jpg"))
	}
	if paths[1] != filepath.Join(wantDir, "a_2.jpg") {
		t.Errorf("second save path = %q, want %q", paths[1], filepath.Join(wantDir, "a_2.jpg"))
	}
}

// TestProcessArticle_SameTitleGetsDistinctDirs 驗證不同文章（不同 URL）
// 標題與推文數相同時，儲存目錄會加上序號後綴而不會互相覆蓋。
func TestProcessArticle_SameTitleGetsDistinctDirs(t *testing.T) {
	c := newCollisionTestCrawler([]string{"http://host1.com/a.jpg"})

	downloadChan := make(chan types.DownloadTask, 10)
	markdownChan := make(chan types.MarkdownInfo, 10)
	ctx := context.Background()

	article1 := types.ArticleInfo{Title: "Same Title", URL: "http://example.com/article1", PushRate: 10}
	article2 := types.ArticleInfo{Title: "Same Title", URL: "http://example.com/article2", PushRate: 10}
	c.processArticle(ctx, article1, downloadChan, markdownChan)
	c.processArticle(ctx, article2, downloadChan, markdownChan)

	close(markdownChan)
	var dirs []string
	for info := range markdownChan {
		dirs = append(dirs, info.SaveDir)
	}

	if len(dirs) != 2 {
		t.Fatalf("expected 2 markdown tasks, got %d", len(dirs))
	}
	if dirs[0] != filepath.Join("test", "Same Title_10") {
		t.Errorf("first dir = %q, want %q", dirs[0], filepath.Join("test", "Same Title_10"))
	}
	if dirs[1] != filepath.Join("test", "Same Title_10_2") {
		t.Errorf("second dir = %q, want %q", dirs[1], filepath.Join("test", "Same Title_10_2"))
	}
}
