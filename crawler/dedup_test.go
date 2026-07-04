package crawler

import (
	"context"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/twtrubiks/ptt-spider-go/config"
	"github.com/twtrubiks/ptt-spider-go/mocks"
	"github.com/twtrubiks/ptt-spider-go/types"
)

func TestUniqueStrings(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{
			name: "無重複",
			in:   []string{"a", "b", "c"},
			want: []string{"a", "b", "c"},
		},
		{
			name: "重複項目只保留首次出現",
			in:   []string{"a", "b", "a", "c", "b"},
			want: []string{"a", "b", "c"},
		},
		{
			name: "空列表",
			in:   nil,
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := uniqueStrings(tt.in); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("uniqueStrings(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

// TestProcessArticle_DeduplicatesImageURLs 驗證同一篇文章中重複的圖片 URL
// 只會派發一次下載任務，避免多個 worker 同時寫入同一檔案造成毀損。
func TestProcessArticle_DeduplicatesImageURLs(t *testing.T) {
	mockClient := &mocks.MockHTTPClient{
		DoFunc: func(_ *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader("<html></html>")),
			}, nil
		},
	}

	// 同一張圖在原文與推文中各出現一次
	mockParser := &mocks.MockParser{
		ParseArticleContentFunc: func(_ io.Reader) (string, []string, error) {
			return "Test Title", []string{
				"http://example.com/img1.jpg",
				"http://example.com/img1.jpg",
				"http://example.com/img2.jpg",
			}, nil
		},
	}

	cfg := &config.Config{
		Crawler: config.CrawlerConfig{
			Delays: config.DelayConfig{MinMs: 0, MaxMs: 0},
		},
	}

	c := NewCrawlerWithDependencies(mockClient, mockParser, mocks.NewMockMarkdownGenerator(), "test", 1, 0, "", cfg)
	progressCh := make(chan types.ProgressEvent, 10)
	c.progress = progressCh

	downloadChan := make(chan types.DownloadTask, 10)
	markdownChan := make(chan types.MarkdownInfo, 10)

	article := types.ArticleInfo{Title: "Test Title", URL: "http://example.com/article", PushRate: 10}
	c.processArticle(context.Background(), article, downloadChan, markdownChan)

	close(downloadChan)
	var tasks []types.DownloadTask
	for task := range downloadChan {
		tasks = append(tasks, task)
	}
	if len(tasks) != 2 {
		t.Errorf("expected 2 download tasks after dedup, got %d", len(tasks))
	}

	close(markdownChan)
	info := <-markdownChan
	if len(info.ImageURLs) != 2 {
		t.Errorf("expected 2 image URLs in markdown info after dedup, got %d", len(info.ImageURLs))
	}

	close(progressCh)
	for evt := range progressCh {
		if evt.Type == types.EventArticleParsed && evt.ImageCount != 2 {
			t.Errorf("EventArticleParsed ImageCount = %d, want 2 (deduped)", evt.ImageCount)
		}
	}
}
