// Package mocks 提供 function field pattern 的 mock 物件，供各套件測試使用。
package mocks

import (
	"io"
	"net/http"
	"strings"

	"github.com/twtrubiks/ptt-spider-go/types"
)

// MockHTTPClient 模擬 HTTP 客戶端
type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

// Do 呼叫 DoFunc（若已設定），否則回傳 200 空回應
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
	ParseMaxPageFunc        func(body io.Reader) (int, error)
}

// ParseArticles 呼叫 ParseArticlesFunc（若已設定），否則回傳空列表
func (m *MockParser) ParseArticles(body io.Reader) ([]types.ArticleInfo, error) {
	if m.ParseArticlesFunc != nil {
		return m.ParseArticlesFunc(body)
	}
	return []types.ArticleInfo{}, nil
}

// ParseArticleContent 呼叫 ParseArticleContentFunc（若已設定），否則回傳固定的測試資料
func (m *MockParser) ParseArticleContent(body io.Reader) (string, []string, error) {
	if m.ParseArticleContentFunc != nil {
		return m.ParseArticleContentFunc(body)
	}
	return "Mock Title", []string{"http://example.com/image.jpg"}, nil
}

// ParseMaxPage 呼叫 ParseMaxPageFunc（若已設定），否則回傳 100
func (m *MockParser) ParseMaxPage(body io.Reader) (int, error) {
	if m.ParseMaxPageFunc != nil {
		return m.ParseMaxPageFunc(body)
	}
	return 100, nil
}

// MockMarkdownGenerator 模擬 Markdown 生成器
type MockMarkdownGenerator struct {
	GenerateFunc func(info types.MarkdownInfo) error
}

// Generate 呼叫 GenerateFunc（若已設定），否則回傳 nil
func (m *MockMarkdownGenerator) Generate(info types.MarkdownInfo) error {
	if m.GenerateFunc != nil {
		return m.GenerateFunc(info)
	}
	return nil
}

// MockLogger 模擬 Logger，可捕捉各等級的日誌呼叫
type MockLogger struct {
	InfoFunc    func(format string, args ...any)
	SuccessFunc func(format string, args ...any)
	ErrorFunc   func(format string, args ...any)
	WarnFunc    func(format string, args ...any)
}

// Info 呼叫 InfoFunc（若已設定）
func (m *MockLogger) Info(format string, args ...any) {
	if m.InfoFunc != nil {
		m.InfoFunc(format, args...)
	}
}

// Success 呼叫 SuccessFunc（若已設定）
func (m *MockLogger) Success(format string, args ...any) {
	if m.SuccessFunc != nil {
		m.SuccessFunc(format, args...)
	}
}

// Error 呼叫 ErrorFunc（若已設定）
func (m *MockLogger) Error(format string, args ...any) {
	if m.ErrorFunc != nil {
		m.ErrorFunc(format, args...)
	}
}

// Warn 呼叫 WarnFunc（若已設定）
func (m *MockLogger) Warn(format string, args ...any) {
	if m.WarnFunc != nil {
		m.WarnFunc(format, args...)
	}
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
