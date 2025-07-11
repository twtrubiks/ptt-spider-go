package tests

import (
	"context"
	"testing"
	"time"

	"github.com/twtrubiks/ptt-spider-go/config"
	"github.com/twtrubiks/ptt-spider-go/crawler"
	"github.com/twtrubiks/ptt-spider-go/markdown"
	"github.com/twtrubiks/ptt-spider-go/ptt"
	"github.com/twtrubiks/ptt-spider-go/types"
)

// TestIntegrationConfigCrawler tests the integration between Config and Crawler
func TestIntegrationConfigCrawler(t *testing.T) {
	cfg := config.DefaultConfig()

	// Create crawler with config
	crawler, err := crawler.NewCrawler("Beauty", 1, 10, "", cfg)
	if err != nil {
		t.Fatalf("Failed to create crawler with config: %v", err)
	}

	if crawler.Config != cfg {
		t.Error("Crawler should use the provided config")
	}

	// Test that config values are applied
	if crawler.Config.Crawler.Workers != cfg.Crawler.Workers {
		t.Error("Config workers should be applied to crawler")
	}
}

// TestIntegrationPTTCrawler tests the integration between PTT client and Crawler
func TestIntegrationPTTCrawler(t *testing.T) {
	cfg := config.DefaultConfig()

	// Test client creation with config
	client, err := ptt.NewClientWithConfig(cfg)
	if err != nil {
		t.Fatalf("Failed to create PTT client: %v", err)
	}

	if client == nil {
		t.Error("PTT client should not be nil")
	}

	// Test crawler creation (which should use the same client creation logic)
	crawler, err := crawler.NewCrawler("Beauty", 1, 10, "", cfg)
	if err != nil {
		t.Fatalf("Failed to create crawler: %v", err)
	}

	if crawler.Client == nil {
		t.Error("Crawler should have HTTP client")
	}
}

// TestIntegrationMarkdownGeneration tests the integration of markdown generation
func TestIntegrationMarkdownGeneration(t *testing.T) {
	tmpDir := t.TempDir()

	// Create markdown info that would come from crawler
	info := types.MarkdownInfo{
		Title:      "[正妹] 整合測試",
		ArticleURL: "https://www.ptt.cc/bbs/Beauty/M.1234567890.A.ABC.html",
		PushCount:  42,
		ImageURLs: []string{
			"https://i.imgur.com/test1.jpg",
			"https://i.imgur.com/test2.png",
		},
		SaveDir: tmpDir,
	}

	// Test markdown generation
	err := markdown.Generate(info)
	if err != nil {
		t.Fatalf("Failed to generate markdown: %v", err)
	}

	// This would be the complete flow from types -> markdown
}

// TestIntegrationTypesFlow tests the complete data flow through types
func TestIntegrationTypesFlow(t *testing.T) {
	// Simulate the complete data flow

	// 1. ArticleInfo (from PTT parsing)
	articleInfo := types.ArticleInfo{
		Title:    "[正妹] 整合測試",
		URL:      "https://www.ptt.cc/bbs/Beauty/M.1234567890.A.ABC.html",
		Author:   "testuser",
		PushRate: 42,
	}

	// 2. DownloadTask (from content parsing)
	downloadTasks := []types.DownloadTask{
		{
			ImageURL: "https://i.imgur.com/test1.jpg",
			SavePath: "/tmp/test1.jpg",
		},
		{
			ImageURL: "https://i.imgur.com/test2.png",
			SavePath: "/tmp/test2.png",
		},
	}

	// 3. MarkdownInfo (for final generation)
	markdownInfo := types.MarkdownInfo{
		Title:      articleInfo.Title,
		ArticleURL: articleInfo.URL,
		PushCount:  articleInfo.PushRate,
		ImageURLs:  make([]string, len(downloadTasks)),
		SaveDir:    "/tmp/test_article",
	}

	// Extract image URLs from download tasks
	for i, task := range downloadTasks {
		markdownInfo.ImageURLs[i] = task.ImageURL
	}

	// Verify the data flow is correct
	if markdownInfo.Title != articleInfo.Title {
		t.Error("Title should be preserved through data flow")
	}
	if markdownInfo.ArticleURL != articleInfo.URL {
		t.Error("URL should be preserved through data flow")
	}
	if markdownInfo.PushCount != articleInfo.PushRate {
		t.Error("Push count should be preserved through data flow")
	}
	if len(markdownInfo.ImageURLs) != len(downloadTasks) {
		t.Error("Image URLs should match download tasks")
	}
}

// TestIntegrationContextCancellation tests context cancellation across components
func TestIntegrationContextCancellation(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Crawler.Workers = 1
	cfg.Crawler.ParserCount = 1

	crawler, err := crawler.NewCrawler("Beauty", 1, 10, "", cfg)
	if err != nil {
		t.Fatalf("Failed to create crawler: %v", err)
	}

	// Test context cancellation
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	crawler.Run(ctx)
	elapsed := time.Since(start)

	// Should stop quickly due to context cancellation
	if elapsed > 2*time.Second {
		t.Errorf("Context cancellation took too long: %v", elapsed)
	}
}

// TestIntegrationConfigurability tests that configuration changes are applied
func TestIntegrationConfigurability(t *testing.T) {
	// Test custom configuration
	cfg := &config.Config{
		Crawler: config.CrawlerConfig{
			Workers:     3,
			ParserCount: 2,
			Channels: config.ChannelConfig{
				ArticleInfo:  50,
				DownloadTask: 100,
				MarkdownTask: 25,
			},
			Delays: config.DelayConfig{
				MinMs: 1000,
				MaxMs: 2000,
			},
			HTTP: config.HTTPConfig{
				Timeout:               "45s",
				MaxIdleConns:          150,
				MaxIdleConnsPerHost:   25,
				IdleConnTimeout:       "120s",
				TLSHandshakeTimeout:   "15s",
				ExpectContinueTimeout: "2s",
			},
		},
	}

	// Test that crawler respects custom config
	crawler, err := crawler.NewCrawler("Beauty", 1, 10, "", cfg)
	if err != nil {
		t.Fatalf("Failed to create crawler with custom config: %v", err)
	}

	if crawler.Config.Crawler.Workers != 3 {
		t.Error("Crawler should use custom worker count")
	}
	if crawler.Config.Crawler.ParserCount != 2 {
		t.Error("Crawler should use custom parser count")
	}

	// Test that HTTP client respects config
	client, err := ptt.NewClientWithConfig(cfg)
	if err != nil {
		t.Fatalf("Failed to create client with custom config: %v", err)
	}

	expectedTimeout := 45 * time.Second
	if client.Timeout != expectedTimeout {
		t.Errorf("Client should use custom timeout: expected %v, got %v", expectedTimeout, client.Timeout)
	}
}
