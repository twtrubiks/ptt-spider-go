package crawler

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/twtrubiks/ptt-spider-go/config"
	"github.com/twtrubiks/ptt-spider-go/mocks"
	"github.com/twtrubiks/ptt-spider-go/types"
	"github.com/twtrubiks/ptt-spider-go/ui"
)

func TestWithProgress_Option(t *testing.T) {
	ch := make(chan types.ProgressEvent, 10)
	cfg := config.DefaultConfig()

	c, err := NewCrawler("test", 1, 0, "", cfg, WithProgress(ch))
	if err != nil {
		t.Fatalf("NewCrawler() error = %v", err)
	}

	if c.progress == nil {
		t.Error("WithProgress should set progress channel")
	}
}

func TestWithLogger_Option(t *testing.T) {
	cfg := config.DefaultConfig()
	logger := ui.NewNoopLogger()

	c, err := NewCrawler("test", 1, 0, "", cfg, WithLogger(logger))
	if err != nil {
		t.Fatalf("NewCrawler() error = %v", err)
	}

	if c.logger != logger {
		t.Error("WithLogger should set custom logger")
	}
}

func TestEmit_NilChannel(t *testing.T) {
	c := &Crawler{} // progress is nil

	// 不應 panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("emit panicked with nil channel: %v", r)
		}
	}()

	c.emit(types.ProgressEvent{Type: types.EventCrawlerDone})
}

func TestEmit_SendsEvent(t *testing.T) {
	ch := make(chan types.ProgressEvent, 1)
	c := &Crawler{progress: ch}

	c.emit(types.ProgressEvent{
		Type:    types.EventCrawlerDone,
		Message: "test",
	})

	select {
	case evt := <-ch:
		if evt.Type != types.EventCrawlerDone {
			t.Errorf("got event type %d, want EventCrawlerDone", evt.Type)
		}
		if evt.Message != "test" {
			t.Errorf("got message %q, want %q", evt.Message, "test")
		}
	default:
		t.Error("expected event on channel, got none")
	}
}

func TestEmit_NonBlocking(t *testing.T) {
	ch := make(chan types.ProgressEvent) // unbuffered, no receiver
	c := &Crawler{progress: ch}

	done := make(chan struct{})
	go func() {
		c.emit(types.ProgressEvent{Type: types.EventCrawlerDone})
		close(done)
	}()

	select {
	case <-done:
		// emit returned without blocking
	case <-time.After(1 * time.Second):
		t.Error("emit blocked on full channel")
	}
}

func TestProgress_ArticleProducerEmitsPageEvents(t *testing.T) {
	progressCh := make(chan types.ProgressEvent, 50)

	mockClient := &mocks.MockHTTPClient{
		DoFunc: func(_ *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader("<html>test</html>")),
			}, nil
		},
	}

	mockParser := &mocks.MockParser{
		ParseMaxPageFunc: func(_ io.Reader) (int, error) {
			return 10, nil
		},
		ParseArticlesFunc: func(_ io.Reader) ([]types.ArticleInfo, error) {
			return []types.ArticleInfo{
				{Title: "Test", URL: "http://example.com/1", PushRate: 20},
			}, nil
		},
	}

	cfg := &config.Config{
		Crawler: config.CrawlerConfig{
			Channels: config.ChannelConfig{ArticleInfo: 10},
		},
	}

	c := NewCrawlerWithDependencies(mockClient, mockParser, mocks.NewMockMarkdownGenerator(), "test", 3, 0, "", cfg)
	c.progress = progressCh

	articleChan := make(chan types.ArticleInfo, 10)
	ctx := context.Background()

	c.articleProducer(ctx, articleChan)

	// 收集 page events
	close(progressCh)
	var pageEvents []types.ProgressEvent
	for evt := range progressCh {
		if evt.Type == types.EventPageParsed {
			pageEvents = append(pageEvents, evt)
		}
	}

	if len(pageEvents) != 3 {
		t.Errorf("expected 3 page events, got %d", len(pageEvents))
	}

	for i, evt := range pageEvents {
		if evt.CurrentPage != i+1 {
			t.Errorf("page event %d: CurrentPage = %d, want %d", i, evt.CurrentPage, i+1)
		}
		if evt.TotalPages != 3 {
			t.Errorf("page event %d: TotalPages = %d, want 3", i, evt.TotalPages)
		}
	}
}

func TestProgress_ProcessArticleEmitsArticleEvent(t *testing.T) {
	progressCh := make(chan types.ProgressEvent, 50)

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
			return "Parsed Title", []string{"http://example.com/img1.jpg", "http://example.com/img2.jpg"}, nil
		},
	}

	cfg := &config.Config{
		Crawler: config.CrawlerConfig{
			Delays: config.DelayConfig{MinMs: 0, MaxMs: 0},
		},
	}

	c := NewCrawlerWithDependencies(mockClient, mockParser, mocks.NewMockMarkdownGenerator(), "test", 1, 0, "", cfg)
	c.progress = progressCh

	downloadChan := make(chan types.DownloadTask, 10)
	markdownChan := make(chan types.MarkdownInfo, 10)
	ctx := context.Background()

	article := types.ArticleInfo{Title: "Original Title", URL: "http://example.com/article", PushRate: 15}
	c.processArticle(ctx, article, downloadChan, markdownChan)

	close(progressCh)
	var articleEvents []types.ProgressEvent
	for evt := range progressCh {
		if evt.Type == types.EventArticleParsed {
			articleEvents = append(articleEvents, evt)
		}
	}

	if len(articleEvents) != 1 {
		t.Fatalf("expected 1 article event, got %d", len(articleEvents))
	}

	evt := articleEvents[0]
	if evt.ArticleTitle != "Original Title" {
		t.Errorf("ArticleTitle = %q, want %q", evt.ArticleTitle, "Original Title")
	}
	if evt.ImageCount != 2 {
		t.Errorf("ImageCount = %d, want 2", evt.ImageCount)
	}
}

func TestProgress_DownloadWorkerEmitsEvents(t *testing.T) {
	progressCh := make(chan types.ProgressEvent, 50)

	mockClient := &mocks.MockHTTPClient{
		DoFunc: func(_ *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 404,
				Body:       io.NopCloser(strings.NewReader("")),
			}, nil
		},
	}

	cfg := &config.Config{
		Crawler: config.CrawlerConfig{
			Delays: config.DelayConfig{MinMs: 0, MaxMs: 0},
		},
	}

	c := NewCrawlerWithDependencies(mockClient, mocks.NewMockParser(), mocks.NewMockMarkdownGenerator(), "test", 1, 0, "", cfg)
	c.progress = progressCh

	taskChan := make(chan types.DownloadTask, 1)
	taskChan <- types.DownloadTask{ImageURL: "http://example.com/img.jpg", SavePath: "/tmp/test.jpg"}
	close(taskChan)

	var wg sync.WaitGroup
	wg.Add(1)
	ctx := context.Background()

	go c.downloadWorker(ctx, 1, taskChan, &wg)
	wg.Wait()

	close(progressCh)
	var startEvents, failEvents []types.ProgressEvent
	for evt := range progressCh {
		switch evt.Type {
		case types.EventDownloadStart:
			startEvents = append(startEvents, evt)
		case types.EventDownloadFail:
			failEvents = append(failEvents, evt)
		}
	}

	if len(startEvents) != 1 {
		t.Errorf("expected 1 download start event, got %d", len(startEvents))
	}
	if len(failEvents) != 1 {
		t.Errorf("expected 1 download fail event, got %d", len(failEvents))
	}
	if len(startEvents) > 0 && startEvents[0].WorkerID != 1 {
		t.Errorf("start event WorkerID = %d, want 1", startEvents[0].WorkerID)
	}
}
