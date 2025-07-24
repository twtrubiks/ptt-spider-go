package ptt

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/twtrubiks/ptt-spider-go/mocks"
)

func TestParserImpl_ParseArticles(t *testing.T) {
	parser := NewParser()

	// Test with valid HTML
	htmlContent := `
		<div class="r-ent">
			<div class="title">
				<a href="/bbs/Beauty/M.1234567890.A.ABC.html">[正妹] 測試標題</a>
			</div>
			<div class="meta">
				<div class="author">testuser</div>
			</div>
			<div class="nrec">
				<span class="hl f1">爆</span>
			</div>
		</div>
		<div class="r-ent">
			<div class="title">
				<a href="/bbs/Beauty/M.1234567891.A.DEF.html">[正妹] 另一個標題</a>
			</div>
			<div class="meta">
				<div class="author">user2</div>
			</div>
			<div class="nrec">
				<span class="hl f3">99</span>
			</div>
		</div>
	`

	articles, err := parser.ParseArticles(strings.NewReader(htmlContent))
	if err != nil {
		t.Fatalf("ParseArticles failed: %v", err)
	}

	if len(articles) != 2 {
		t.Errorf("Expected 2 articles, got %d", len(articles))
	}

	// Check first article
	if articles[0].Title != "[正妹] 測試標題" {
		t.Errorf("Expected title '[正妹] 測試標題', got %s", articles[0].Title)
	}
	if articles[0].PushRate != 100 {
		t.Errorf("Expected push rate 100, got %d", articles[0].PushRate)
	}
	if articles[0].Author != "testuser" {
		t.Errorf("Expected author 'testuser', got %s", articles[0].Author)
	}

	// Check second article
	if articles[1].Title != "[正妹] 另一個標題" {
		t.Errorf("Expected title '[正妹] 另一個標題', got %s", articles[1].Title)
	}
	if articles[1].PushRate != 99 {
		t.Errorf("Expected push rate 99, got %d", articles[1].PushRate)
	}

	// Test with empty HTML (goquery doesn't error on invalid HTML, just returns empty results)
	articles, err = parser.ParseArticles(strings.NewReader(""))
	if err != nil {
		t.Errorf("ParseArticles should not error on empty HTML: %v", err)
	}
	if len(articles) != 0 {
		t.Errorf("Expected 0 articles for empty HTML, got %d", len(articles))
	}
}

func TestParserImpl_ParseArticleContent(t *testing.T) {
	parser := NewParser()

	// Test with valid article HTML
	htmlContent := `
		<div class="article-meta-tag">標題</div>
		<div class="article-meta-value">[正妹] 測試文章標題</div>
		<div id="main-content">
			<a href="https://i.imgur.com/test1.jpg">Image 1</a>
			<a href="//i.imgur.com/test2.png">Image 2</a>
			<a href="http://example.com/test3.gif">Image 3</a>
			<a href="https://imgur.com/abcd123">Imgur without extension</a>
			<a href="https://imgur.com/a/album123">Imgur album (should be ignored)</a>
		</div>
	`

	title, images, err := parser.ParseArticleContent(strings.NewReader(htmlContent))
	if err != nil {
		t.Fatalf("ParseArticleContent failed: %v", err)
	}

	if title != "[正妹] 測試文章標題" {
		t.Errorf("Expected title '[正妹] 測試文章標題', got '%s'", title)
	}

	expectedImages := 4 // jpg, png, gif (converted from http to https), and imgur without extension
	if len(images) != expectedImages {
		t.Errorf("Expected %d images, got %d: %v", expectedImages, len(images), images)
	}

	// Check specific image conversions
	found := make(map[string]bool)
	for _, img := range images {
		found[img] = true
	}

	if !found["https://i.imgur.com/test1.jpg"] {
		t.Error("Expected https://i.imgur.com/test1.jpg in images")
	}
	if !found["https://i.imgur.com/test2.png"] {
		t.Error("Expected https://i.imgur.com/test2.png in images")
	}
	if !found["https://example.com/test3.gif"] {
		t.Error("Expected https://example.com/test3.gif in images (http converted to https)")
	}
	if !found["https://imgur.com/abcd123.jpg"] {
		t.Error("Expected https://imgur.com/abcd123.jpg in images (imgur without extension)")
	}

	// Test with empty HTML
	title, images, err = parser.ParseArticleContent(strings.NewReader(""))
	if err != nil {
		t.Errorf("ParseArticleContent should not error on empty HTML: %v", err)
	}
	if title != "" {
		t.Errorf("Expected empty title for empty HTML, got '%s'", title)
	}
	if len(images) != 0 {
		t.Errorf("Expected 0 images for empty HTML, got %d", len(images))
	}
}

func TestParserImpl_GetMaxPage(t *testing.T) {
	parser := NewParser()

	// Test successful response
	mockClient := &mocks.MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			htmlContent := `
				<div class="btn-group-paging">
					<a href="/bbs/Beauty/index2345.html">‹ 上頁</a>
				</div>
			`
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(htmlContent)),
			}, nil
		},
	}

	maxPage, err := parser.GetMaxPage(context.Background(), mockClient, "Beauty")
	if err != nil {
		t.Fatalf("GetMaxPage failed: %v", err)
	}

	if maxPage != 2346 { // 2345 + 1
		t.Errorf("Expected max page 2346, got %d", maxPage)
	}

	// Test network error
	mockClient.DoFunc = func(req *http.Request) (*http.Response, error) {
		return nil, fmt.Errorf("network error")
	}

	_, err = parser.GetMaxPage(context.Background(), mockClient, "Beauty")
	if err == nil {
		t.Error("Expected error for network failure, got nil")
	}

	// Test HTTP error status
	mockClient.DoFunc = func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 404,
			Body:       io.NopCloser(strings.NewReader("")),
		}, nil
	}

	_, err = parser.GetMaxPage(context.Background(), mockClient, "Beauty")
	if err == nil {
		t.Error("Expected error for 404 status, got nil")
	}

	// Test missing previous page button
	mockClient.DoFunc = func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader("<div>No previous page button</div>")),
		}, nil
	}

	_, err = parser.GetMaxPage(context.Background(), mockClient, "Beauty")
	if err == nil {
		t.Error("Expected error for missing previous page button, got nil")
	}

	// Test invalid page number format
	mockClient.DoFunc = func(req *http.Request) (*http.Response, error) {
		htmlContent := `
			<div class="btn-group-paging">
				<a href="/bbs/Beauty/invalid.html">‹ 上頁</a>
			</div>
		`
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(htmlContent)),
		}, nil
	}

	_, err = parser.GetMaxPage(context.Background(), mockClient, "Beauty")
	if err == nil {
		t.Error("Expected error for invalid page format, got nil")
	}
}

func TestNewParser(t *testing.T) {
	parser := NewParser()
	if parser == nil {
		t.Error("NewParser should return a valid parser instance")
	}

	// Check that it implements the Parser interface by calling methods
	_, err := parser.ParseArticles(strings.NewReader(""))
	if err != nil {
		t.Errorf("ParseArticles should not error on empty HTML: %v", err)
	}
}
