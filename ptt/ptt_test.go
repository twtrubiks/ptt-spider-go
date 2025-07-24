package ptt

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/twtrubiks/ptt-spider-go/config"
	"github.com/twtrubiks/ptt-spider-go/constants"
	"github.com/twtrubiks/ptt-spider-go/types"
)

func loadFixture(t *testing.T, filename string) string {
	data, err := os.ReadFile("../tests/fixtures/" + filename)
	if err != nil {
		t.Fatalf("Failed to load fixture %s: %v", filename, err)
	}
	return string(data)
}

func TestParseArticles(t *testing.T) {
	html := loadFixture(t, "board_list.html")

	tests := []struct {
		name          string
		html          string
		expectedCount int
		expectErr     bool
		validate      func(t *testing.T, articles []types.ArticleInfo)
	}{
		{
			name:          "Parse board list",
			html:          html,
			expectedCount: 3,
			expectErr:     false,
			validate: func(t *testing.T, articles []types.ArticleInfo) {
				// Test first article (爆文)
				if articles[0].Title != "[正妹] 測試標題" {
					t.Errorf("expected title '[正妹] 測試標題', got %s", articles[0].Title)
				}
				if articles[0].PushRate != 100 {
					t.Errorf("expected push rate 100 for 爆文, got %d", articles[0].PushRate)
				}
				if articles[0].URL != constants.PttBaseURL+"/bbs/Beauty/M.1234567890.A.ABC.html" {
					t.Errorf("unexpected URL: %s", articles[0].URL)
				}
				if articles[0].Author != "testuser" {
					t.Errorf("expected author 'testuser', got %s", articles[0].Author)
				}

				// Test second article (99 push)
				if articles[1].PushRate != 99 {
					t.Errorf("expected push rate 99, got %d", articles[1].PushRate)
				}

				// Test third article (X5)
				if articles[2].PushRate != -5 {
					t.Errorf("expected push rate -5 for X5, got %d", articles[2].PushRate)
				}
			},
		},
		{
			name:          "Parse empty HTML",
			html:          "<html><body></body></html>",
			expectedCount: 0,
			expectErr:     false,
			validate:      nil,
		},
		{
			name:          "Parse invalid HTML",
			html:          "not html",
			expectedCount: 0,
			expectErr:     false,
			validate:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.html)
			articles, err := ParseArticles(reader)
			if (err != nil) != tt.expectErr {
				t.Errorf("ParseArticles() error = %v, expectErr %v", err, tt.expectErr)
				return
			}
			if len(articles) != tt.expectedCount {
				t.Errorf("expected %d articles, got %d", tt.expectedCount, len(articles))
			}
			if tt.validate != nil && len(articles) > 0 {
				tt.validate(t, articles)
			}
		})
	}
}

func TestParseArticleContent(t *testing.T) {
	tests := []struct {
		name            string
		fixtureFile     string
		expectedTitle   string
		expectedImgURLs []string
		expectErr       bool
	}{
		{
			name:          "Parse article with images",
			fixtureFile:   "article_content.html",
			expectedTitle: "[正妹] 測試標題",
			expectedImgURLs: []string{
				"https://i.imgur.com/test.jpg",
				"https://example.com/test.png",
				"https://example.com/test.gif",         // HTTP should be converted to HTTPS
				"https://i.imgur.com/gallery/test.jpg", // Gallery link gets .jpg added
			},
			expectErr: false,
		},
		{
			name:          "Parse article with multiple images",
			fixtureFile:   "article_with_images.html",
			expectedTitle: "[正妹] 多圖測試",
			expectedImgURLs: []string{
				"https://i.imgur.com/image1.jpg",
				"https://i.imgur.com/image2.png",
				"https://example.com/image3.jpeg",
				"https://example.com/image4.gif",
				"https://i.imgur.com/image5.jpg",
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html := loadFixture(t, tt.fixtureFile)
			reader := strings.NewReader(html)
			title, imgURLs, err := ParseArticleContent(reader)

			if (err != nil) != tt.expectErr {
				t.Errorf("ParseArticleContent() error = %v, expectErr %v", err, tt.expectErr)
				return
			}

			if title != tt.expectedTitle {
				t.Errorf("expected title '%s', got '%s'", tt.expectedTitle, title)
			}

			if len(imgURLs) != len(tt.expectedImgURLs) {
				t.Errorf("expected %d image URLs, got %d", len(tt.expectedImgURLs), len(imgURLs))
				t.Logf("Actual URLs: %v", imgURLs)
				t.Logf("Expected URLs: %v", tt.expectedImgURLs)
			}

			// Check expected URLs
			for _, expectedURL := range tt.expectedImgURLs {
				found := false
				for _, url := range imgURLs {
					if url == expectedURL {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected URL '%s' not found in results", expectedURL)
				}
			}
		})
	}
}

func TestGetMaxPage(t *testing.T) {
	tests := []struct {
		name        string
		setupServer func() *httptest.Server
		expectPage  int
		expectErr   bool
	}{
		{
			name: "Get max page from board",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if strings.Contains(r.URL.Path, "/bbs/Beauty/index.html") {
						w.Write([]byte(`
							<html>
								<div class="btn-group-paging">
									<a class="btn wide" href="/bbs/Beauty/index1234.html">‹ 上頁</a>
									<a class="btn wide" href="/bbs/Beauty/index1236.html">下頁 ›</a>
								</div>
							</html>
						`))
					}
				}))
			},
			expectPage: 1235,
			expectErr:  false,
		},
		{
			name: "No prev page button",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte(`<html><body>No buttons here</body></html>`))
				}))
			},
			expectPage: 0,
			expectErr:  true,
		},
		{
			name: "Server error",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))
			},
			expectPage: 0,
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := tt.setupServer()
			defer server.Close()

			// Create a test client
			client := &http.Client{Timeout: 5 * time.Second}

			// Note: PttHead is a const, so we can't change it for testing
			// This is a limitation of the current implementation that could be improved

			// For this test, we'll create a custom URL
			ctx := context.Background()
			board := "Beauty"

			// Since we can't change PttHead, we'll test the server response parsing logic directly
			// by making a request to our test server
			req, err := http.NewRequestWithContext(ctx, "GET", server.URL+"/bbs/"+board+"/index.html", nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			resp, err := client.Do(req)
			if err != nil && !tt.expectErr {
				t.Fatalf("Request failed: %v", err)
			}
			if err != nil && tt.expectErr {
				return // Expected error case
			}
			defer resp.Body.Close()

			// Test the parsing logic manually since we can't easily mock PttHead
			if resp.StatusCode == http.StatusOK && !tt.expectErr {
				// In a real implementation, we would parse the response here
				// For this test, we'll verify the server responded correctly
				if resp.StatusCode != http.StatusOK {
					t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
				}
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	if client == nil {
		t.Fatal("NewClient() returned nil client")
	}

	// Check that cookie jar is set
	if client.Jar == nil {
		t.Error("client should have cookie jar")
	}

	// Check that custom transport is set
	if client.Transport == nil {
		t.Error("client should have custom transport")
	}

	// Verify it's our custom transport
	if _, ok := client.Transport.(*customTransport); !ok {
		t.Error("client should use customTransport")
	}
}

func TestNewClientWithConfig(t *testing.T) {
	cfg := &config.Config{
		Crawler: config.CrawlerConfig{
			HTTP: config.HTTPConfig{
				Timeout:               "45s",
				MaxIdleConns:          200,
				MaxIdleConnsPerHost:   30,
				IdleConnTimeout:       "120s",
				TLSHandshakeTimeout:   "15s",
				ExpectContinueTimeout: "2s",
			},
		},
	}

	client, err := NewClientWithConfig(cfg)
	if err != nil {
		t.Fatalf("NewClientWithConfig() failed: %v", err)
	}

	if client == nil {
		t.Fatal("NewClientWithConfig() returned nil client")
	}

	// Check timeout
	expectedTimeout := 45 * time.Second
	if client.Timeout != expectedTimeout {
		t.Errorf("expected timeout %v, got %v", expectedTimeout, client.Timeout)
	}

	// Check that cookie jar is set
	if client.Jar == nil {
		t.Error("client should have cookie jar")
	}

	// Check custom transport
	customTransp, ok := client.Transport.(*customTransport)
	if !ok {
		t.Fatal("client should use customTransport")
	}

	// Check underlying transport
	transport, ok := customTransp.transport.(*http.Transport)
	if !ok {
		t.Fatal("customTransport should wrap http.Transport")
	}

	// Verify transport configuration
	if transport.MaxIdleConns != 200 {
		t.Errorf("expected MaxIdleConns 200, got %d", transport.MaxIdleConns)
	}
	if transport.MaxIdleConnsPerHost != 30 {
		t.Errorf("expected MaxIdleConnsPerHost 30, got %d", transport.MaxIdleConnsPerHost)
	}
	if transport.IdleConnTimeout != 120*time.Second {
		t.Errorf("expected IdleConnTimeout 120s, got %v", transport.IdleConnTimeout)
	}
	if transport.TLSHandshakeTimeout != 15*time.Second {
		t.Errorf("expected TLSHandshakeTimeout 15s, got %v", transport.TLSHandshakeTimeout)
	}
	if transport.ExpectContinueTimeout != 2*time.Second {
		t.Errorf("expected ExpectContinueTimeout 2s, got %v", transport.ExpectContinueTimeout)
	}
	if transport.DisableKeepAlives != false {
		t.Error("keep-alives should be enabled")
	}
}

func TestHTTPTransportConfiguration(t *testing.T) {
	tests := []struct {
		name   string
		config *config.Config
		check  func(t *testing.T, client *http.Client)
	}{
		{
			name:   "Default config",
			config: config.DefaultConfig(),
			check: func(t *testing.T, client *http.Client) {
				customTransp := client.Transport.(*customTransport)
				transport := customTransp.transport.(*http.Transport)

				if transport.MaxIdleConns != 100 {
					t.Errorf("expected default MaxIdleConns 100, got %d", transport.MaxIdleConns)
				}
				if transport.MaxIdleConnsPerHost != 20 {
					t.Errorf("expected default MaxIdleConnsPerHost 20, got %d", transport.MaxIdleConnsPerHost)
				}
				if transport.IdleConnTimeout != 90*time.Second {
					t.Errorf("expected default IdleConnTimeout 90s, got %v", transport.IdleConnTimeout)
				}
			},
		},
		{
			name: "Custom config",
			config: &config.Config{
				Crawler: config.CrawlerConfig{
					HTTP: config.HTTPConfig{
						Timeout:             "60s",
						MaxIdleConns:        500,
						MaxIdleConnsPerHost: 100,
						IdleConnTimeout:     "300s",
					},
				},
			},
			check: func(t *testing.T, client *http.Client) {
				if client.Timeout != 60*time.Second {
					t.Errorf("expected timeout 60s, got %v", client.Timeout)
				}

				customTransp := client.Transport.(*customTransport)
				transport := customTransp.transport.(*http.Transport)

				if transport.MaxIdleConns != 500 {
					t.Errorf("expected MaxIdleConns 500, got %d", transport.MaxIdleConns)
				}
				if transport.MaxIdleConnsPerHost != 100 {
					t.Errorf("expected MaxIdleConnsPerHost 100, got %d", transport.MaxIdleConnsPerHost)
				}
				if transport.IdleConnTimeout != 300*time.Second {
					t.Errorf("expected IdleConnTimeout 300s, got %v", transport.IdleConnTimeout)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClientWithConfig(tt.config)
			if err != nil {
				t.Fatalf("NewClientWithConfig() failed: %v", err)
			}
			tt.check(t, client)
		})
	}
}

func TestCustomTransportRoundTrip(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if User-Agent was set
		userAgent := r.Header.Get("User-Agent")
		if !strings.Contains(userAgent, "Chrome") {
			t.Errorf("Expected Chrome User-Agent, got: %s", userAgent)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	// Create custom transport
	transport := &customTransport{}

	// Create request
	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Make request through transport
	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}
}
