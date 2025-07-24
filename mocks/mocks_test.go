package mocks

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/twtrubiks/ptt-spider-go/interfaces"
	"github.com/twtrubiks/ptt-spider-go/types"
)

func TestMockHTTPClient(t *testing.T) {
	// Test default behavior
	client := NewMockHTTPClient()
	req, _ := http.NewRequest("GET", "http://example.com", nil)

	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Test custom behavior
	client.DoFunc = func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 404,
			Body:       io.NopCloser(strings.NewReader("Not Found")),
		}, nil
	}

	resp, err = client.Do(req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if resp.StatusCode != 404 {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}

	// Test error behavior
	client.DoFunc = func(req *http.Request) (*http.Response, error) {
		return nil, fmt.Errorf("connection failed")
	}

	_, err = client.Do(req)
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if err.Error() != "connection failed" {
		t.Errorf("Expected 'connection failed', got '%s'", err.Error())
	}
}

func TestMockParser(t *testing.T) {
	parser := NewMockParser()

	t.Run("default ParseArticles behavior", func(t *testing.T) {
		testDefaultParseArticles(t, parser)
	})

	t.Run("custom ParseArticles behavior", func(t *testing.T) {
		testCustomParseArticles(t, parser)
	})

	t.Run("default ParseArticleContent behavior", func(t *testing.T) {
		testDefaultParseArticleContent(t, parser)
	})

	t.Run("custom ParseArticleContent behavior", func(t *testing.T) {
		testCustomParseArticleContent(t, parser)
	})

	t.Run("default GetMaxPage behavior", func(t *testing.T) {
		testDefaultGetMaxPage(t, parser)
	})

	t.Run("custom GetMaxPage behavior", func(t *testing.T) {
		testCustomGetMaxPage(t, parser)
	})
}

func testDefaultParseArticles(t *testing.T, parser *MockParser) {
	articles, err := parser.ParseArticles(strings.NewReader(""))
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(articles) != 0 {
		t.Errorf("Expected empty articles, got %v", articles)
	}
}

func testCustomParseArticles(t *testing.T, parser *MockParser) {
	parser.ParseArticlesFunc = func(body io.Reader) ([]types.ArticleInfo, error) {
		return []types.ArticleInfo{
			{Title: "Test Article", URL: "http://example.com", PushRate: 10},
		}, nil
	}

	articles, err := parser.ParseArticles(strings.NewReader(""))
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(articles) != 1 {
		t.Errorf("Expected 1 article, got %d", len(articles))
	}
	if articles[0].Title != "Test Article" {
		t.Errorf("Expected 'Test Article', got '%s'", articles[0].Title)
	}
}

func testDefaultParseArticleContent(t *testing.T, parser *MockParser) {
	title, images, err := parser.ParseArticleContent(strings.NewReader(""))
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if title != "Mock Title" {
		t.Errorf("Expected 'Mock Title', got '%s'", title)
	}
	if len(images) != 1 || images[0] != "http://example.com/image.jpg" {
		t.Errorf("Expected default image, got %v", images)
	}
}

func testCustomParseArticleContent(t *testing.T, parser *MockParser) {
	parser.ParseArticleContentFunc = func(body io.Reader) (string, []string, error) {
		return "Custom Title", []string{"http://custom.com/image1.jpg", "http://custom.com/image2.jpg"}, nil
	}

	title, images, err := parser.ParseArticleContent(strings.NewReader(""))
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if title != "Custom Title" {
		t.Errorf("Expected 'Custom Title', got '%s'", title)
	}
	if len(images) != 2 {
		t.Errorf("Expected 2 images, got %d", len(images))
	}
}

func testDefaultGetMaxPage(t *testing.T, parser *MockParser) {
	maxPage, err := parser.GetMaxPage(context.Background(), nil, "testboard")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if maxPage != 100 {
		t.Errorf("Expected 100, got %d", maxPage)
	}
}

func testCustomGetMaxPage(t *testing.T, parser *MockParser) {
	parser.GetMaxPageFunc = func(ctx context.Context, client interfaces.HTTPClient, board string) (int, error) {
		return 50, nil
	}

	maxPage, err := parser.GetMaxPage(context.Background(), nil, "testboard")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if maxPage != 50 {
		t.Errorf("Expected 50, got %d", maxPage)
	}
}

func TestMockMarkdownGenerator(t *testing.T) {
	generator := NewMockMarkdownGenerator()

	// Test default behavior
	info := types.MarkdownInfo{
		Title:      "Test Article",
		ArticleURL: "http://example.com",
		PushCount:  10,
		ImageURLs:  []string{"http://example.com/image.jpg"},
		SaveDir:    "/tmp/test",
	}

	err := generator.Generate(info)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test custom behavior
	generator.GenerateFunc = func(info types.MarkdownInfo) error {
		if info.Title != "Test Article" {
			return fmt.Errorf("unexpected title: %s", info.Title)
		}
		return nil
	}

	err = generator.Generate(info)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test error behavior
	generator.GenerateFunc = func(info types.MarkdownInfo) error {
		return fmt.Errorf("generation failed")
	}

	err = generator.Generate(info)
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if err.Error() != "generation failed" {
		t.Errorf("Expected 'generation failed', got '%s'", err.Error())
	}
}
