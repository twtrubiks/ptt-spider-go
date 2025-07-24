package mocks

import (
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/twtrubiks/ptt-spider-go/interfaces"
	"github.com/twtrubiks/ptt-spider-go/types"
)

// MockHTTPClient 模擬 HTTP 客戶端
type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if m.DoFunc != nil {
		return m.DoFunc(req)
	}
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader("")),
	}, nil
}

// MockParser 模擬解析器
type MockParser struct {
	ParseArticlesFunc       func(body io.Reader) ([]types.ArticleInfo, error)
	ParseArticleContentFunc func(body io.Reader) (string, []string, error)
	GetMaxPageFunc          func(ctx context.Context, client interfaces.HTTPClient, board string) (int, error)
}

func (m *MockParser) ParseArticles(body io.Reader) ([]types.ArticleInfo, error) {
	if m.ParseArticlesFunc != nil {
		return m.ParseArticlesFunc(body)
	}
	return []types.ArticleInfo{}, nil
}

func (m *MockParser) ParseArticleContent(body io.Reader) (string, []string, error) {
	if m.ParseArticleContentFunc != nil {
		return m.ParseArticleContentFunc(body)
	}
	return "Mock Title", []string{"http://example.com/image.jpg"}, nil
}

func (m *MockParser) GetMaxPage(ctx context.Context, client interfaces.HTTPClient, board string) (int, error) {
	if m.GetMaxPageFunc != nil {
		return m.GetMaxPageFunc(ctx, client, board)
	}
	return 100, nil
}

// MockMarkdownGenerator 模擬 Markdown 生成器
type MockMarkdownGenerator struct {
	GenerateFunc func(info types.MarkdownInfo) error
}

func (m *MockMarkdownGenerator) Generate(info types.MarkdownInfo) error {
	if m.GenerateFunc != nil {
		return m.GenerateFunc(info)
	}
	return nil
}

// NewMockHTTPClient 建立新的模擬 HTTP 客戶端
func NewMockHTTPClient() *MockHTTPClient {
	return &MockHTTPClient{}
}

// NewMockParser 建立新的模擬解析器
func NewMockParser() *MockParser {
	return &MockParser{}
}

// NewMockMarkdownGenerator 建立新的模擬 Markdown 生成器
func NewMockMarkdownGenerator() *MockMarkdownGenerator {
	return &MockMarkdownGenerator{}
}
