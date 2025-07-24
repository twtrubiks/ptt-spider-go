package crawler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/twtrubiks/ptt-spider-go/config"
	"github.com/twtrubiks/ptt-spider-go/interfaces"
	"github.com/twtrubiks/ptt-spider-go/mocks"
	"github.com/twtrubiks/ptt-spider-go/types"
)

func TestNewCrawlerWithDependencies(t *testing.T) {
	mockClient := mocks.NewMockHTTPClient()
	mockParser := mocks.NewMockParser()
	mockMarkdownGen := mocks.NewMockMarkdownGenerator()

	cfg := &config.Config{
		Crawler: config.CrawlerConfig{
			Workers:     2,
			ParserCount: 1,
			Channels: config.ChannelConfig{
				ArticleInfo:  10,
				DownloadTask: 50,
				MarkdownTask: 10,
			},
		},
	}

	crawler := NewCrawlerWithDependencies(
		mockClient,
		mockParser,
		mockMarkdownGen,
		"testboard",
		3,
		10,
		"",
		cfg,
	)

	// Verify all fields are set correctly
	if crawler.Client != mockClient {
		t.Error("Client should be set to mock client")
	}
	if crawler.Parser != mockParser {
		t.Error("Parser should be set to mock parser")
	}
	if crawler.MarkdownGenerator != mockMarkdownGen {
		t.Error("MarkdownGenerator should be set to mock generator")
	}
	if crawler.Board != "testboard" {
		t.Errorf("Board = %s, want testboard", crawler.Board)
	}
	if crawler.Pages != 3 {
		t.Errorf("Pages = %d, want 3", crawler.Pages)
	}
	if crawler.PushRate != 10 {
		t.Errorf("PushRate = %d, want 10", crawler.PushRate)
	}
}

func TestCrawler_InitializeChannels(t *testing.T) {
	cfg := &config.Config{
		Crawler: config.CrawlerConfig{
			Channels: config.ChannelConfig{
				ArticleInfo:  15,
				DownloadTask: 25,
				MarkdownTask: 5,
			},
		},
	}

	crawler := &Crawler{Config: cfg}
	channels := crawler.initializeChannels()

	// Check channel capacities
	if cap(channels.ArticleInfo) != 15 {
		t.Errorf("ArticleInfo channel capacity = %d, want 15", cap(channels.ArticleInfo))
	}
	if cap(channels.DownloadTask) != 25 {
		t.Errorf("DownloadTask channel capacity = %d, want 25", cap(channels.DownloadTask))
	}
	if cap(channels.MarkdownTask) != 5 {
		t.Errorf("MarkdownTask channel capacity = %d, want 5", cap(channels.MarkdownTask))
	}
}

func TestCrawler_ArticleProducerWithMocks(t *testing.T) {
	mockClient := &mocks.MockHTTPClient{}
	mockParser := &mocks.MockParser{}

	// Set up mock parser to return test data
	mockParser.GetMaxPageFunc = func(ctx context.Context, client interfaces.HTTPClient, board string) (int, error) {
		return 5, nil
	}

	mockParser.ParseArticlesFunc = func(body io.Reader) ([]types.ArticleInfo, error) {
		return []types.ArticleInfo{
			{Title: "Test Article 1", URL: "http://example.com/1", PushRate: 15},
			{Title: "Test Article 2", URL: "http://example.com/2", PushRate: 20},
		}, nil
	}

	// Set up mock client to return test HTML
	mockClient.DoFunc = func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader("<html>test</html>")),
		}, nil
	}

	cfg := &config.Config{
		Crawler: config.CrawlerConfig{
			Channels: config.ChannelConfig{
				ArticleInfo: 10,
			},
		},
	}

	crawler := NewCrawlerWithDependencies(
		mockClient,
		mockParser,
		mocks.NewMockMarkdownGenerator(),
		"testboard",
		2, // 2 pages
		10,
		"",
		cfg,
	)

	// Test article producer
	articleChan := make(chan types.ArticleInfo, 10)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go crawler.articleProducer(ctx, articleChan)

	// Collect articles
	var articles []types.ArticleInfo
	for article := range articleChan {
		articles = append(articles, article)
	}

	// We expect articles from 2 pages, each page has 2 articles with pushRate >= 10
	expectedCount := 4 // 2 pages * 2 articles per page
	if len(articles) != expectedCount {
		t.Errorf("Expected %d articles, got %d", expectedCount, len(articles))
	}

	// Check that only articles with sufficient push rate are included
	for _, article := range articles {
		if article.PushRate < 10 {
			t.Errorf("Article with push rate %d should be filtered out", article.PushRate)
		}
	}
}

func TestCrawler_ContentParserWithMocks(t *testing.T) {
	mockClient := &mocks.MockHTTPClient{}
	mockParser := &mocks.MockParser{}
	mockMarkdownGen := &mocks.MockMarkdownGenerator{}

	// Set up mock parser
	mockParser.ParseArticleContentFunc = func(body io.Reader) (string, []string, error) {
		return "Parsed Title", []string{"http://example.com/image1.jpg", "http://example.com/image2.jpg"}, nil
	}

	// Set up mock client
	mockClient.DoFunc = func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader("<html>article content</html>")),
		}, nil
	}

	// Track markdown generation calls
	mockMarkdownGen.GenerateFunc = func(info types.MarkdownInfo) error {
		return nil
	}

	cfg := &config.Config{
		Crawler: config.CrawlerConfig{
			Workers:     2,
			ParserCount: 1,
			Delays: config.DelayConfig{
				MinMs: 100,
				MaxMs: 200,
			},
		},
	}

	crawler := NewCrawlerWithDependencies(
		mockClient,
		mockParser,
		mockMarkdownGen,
		"testboard",
		1,
		10,
		"",
		cfg,
	)

	// Create channels
	articleChan := make(chan types.ArticleInfo, 1)
	downloadChan := make(chan types.DownloadTask, 5)
	markdownChan := make(chan types.MarkdownInfo, 1)

	// Create waitgroups
	var wg sync.WaitGroup
	wg.Add(1) // Add one for the content parser

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Send test article
	testArticle := types.ArticleInfo{
		Title:    "Original Title",
		URL:      "http://example.com/article",
		PushRate: 15,
	}
	articleChan <- testArticle
	close(articleChan)

	// Start content parser
	go crawler.contentParser(ctx, &wg, articleChan, downloadChan, markdownChan)

	// Wait for processing to complete
	wg.Wait()
	close(downloadChan)
	close(markdownChan)

	// Check download tasks
	var downloadTasks []types.DownloadTask
	for task := range downloadChan {
		downloadTasks = append(downloadTasks, task)
	}

	if len(downloadTasks) != 2 {
		t.Errorf("Expected 2 download tasks, got %d", len(downloadTasks))
	}

	// Check markdown tasks
	var markdownTasks []types.MarkdownInfo
	for task := range markdownChan {
		markdownTasks = append(markdownTasks, task)
	}

	if len(markdownTasks) != 1 {
		t.Errorf("Expected 1 markdown task, got %d", len(markdownTasks))
	}

	if len(markdownTasks) > 0 {
		task := markdownTasks[0]
		if task.Title != "Original Title" {
			t.Errorf("Expected title 'Original Title', got '%s'", task.Title)
		}
		if len(task.ImageURLs) != 2 {
			t.Errorf("Expected 2 image URLs, got %d", len(task.ImageURLs))
		}
	}
}

func TestCrawler_MarkdownWorkerWithMocks(t *testing.T) {
	mockMarkdownGen := &mocks.MockMarkdownGenerator{}

	// Track generation calls
	var generatedInfos []types.MarkdownInfo
	mockMarkdownGen.GenerateFunc = func(info types.MarkdownInfo) error {
		generatedInfos = append(generatedInfos, info)
		return nil
	}

	crawler := NewCrawlerWithDependencies(
		mocks.NewMockHTTPClient(),
		mocks.NewMockParser(),
		mockMarkdownGen,
		"testboard",
		1,
		10,
		"",
		&config.Config{},
	)

	// Create test channel
	markdownChan := make(chan types.MarkdownInfo, 2)
	var wg sync.WaitGroup

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start markdown worker
	wg.Add(1)
	go crawler.markdownWorker(ctx, markdownChan, &wg)

	// Send test tasks
	testInfo1 := types.MarkdownInfo{Title: "Article 1", ArticleURL: "http://example.com/1"}
	testInfo2 := types.MarkdownInfo{Title: "Article 2", ArticleURL: "http://example.com/2"}

	markdownChan <- testInfo1
	markdownChan <- testInfo2
	close(markdownChan)

	// Wait for processing
	time.Sleep(500 * time.Millisecond)

	// Check results
	if len(generatedInfos) != 2 {
		t.Errorf("Expected 2 generated infos, got %d", len(generatedInfos))
	}

	if len(generatedInfos) >= 1 && generatedInfos[0].Title != "Article 1" {
		t.Errorf("Expected first article title 'Article 1', got '%s'", generatedInfos[0].Title)
	}
}

func TestCrawler_ErrorHandling(t *testing.T) {
	mockClient := &mocks.MockHTTPClient{}
	mockParser := &mocks.MockParser{}

	// Set up mock to return error
	mockClient.DoFunc = func(req *http.Request) (*http.Response, error) {
		return nil, fmt.Errorf("network error")
	}

	cfg := &config.Config{
		Crawler: config.CrawlerConfig{
			Channels: config.ChannelConfig{
				ArticleInfo: 1,
			},
		},
	}

	crawler := NewCrawlerWithDependencies(
		mockClient,
		mockParser,
		mocks.NewMockMarkdownGenerator(),
		"testboard",
		1,
		10,
		"",
		cfg,
	)

	// Test that articleProducer handles errors gracefully
	articleChan := make(chan types.ArticleInfo, 1)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// This should complete without panicking even with network errors
	crawler.articleProducer(ctx, articleChan)

	// Channel should be closed (no articles due to error)
	select {
	case _, ok := <-articleChan:
		if ok {
			t.Error("Expected channel to be closed due to error")
		}
	default:
		// Channel is closed, which is expected
	}
}
