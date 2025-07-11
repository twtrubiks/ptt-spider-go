package crawler

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/twtrubiks/ptt-spider-go/config"
)

func TestNewCrawler(t *testing.T) {
	cfg := config.DefaultConfig()

	tests := []struct {
		name     string
		board    string
		pages    int
		pushRate int
		fileURL  string
		wantErr  bool
	}{
		{
			name:     "Valid crawler creation",
			board:    "Beauty",
			pages:    3,
			pushRate: 10,
			fileURL:  "",
			wantErr:  false,
		},
		{
			name:     "Crawler with file mode",
			board:    "",
			pages:    0,
			pushRate: 0,
			fileURL:  "test.txt",
			wantErr:  false,
		},
		{
			name:     "Empty parameters",
			board:    "",
			pages:    0,
			pushRate: 0,
			fileURL:  "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			crawler, err := NewCrawler(tt.board, tt.pages, tt.pushRate, tt.fileURL, cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCrawler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if crawler == nil {
					t.Error("NewCrawler() returned nil crawler")
					return
				}
				if crawler.Board != tt.board {
					t.Errorf("Expected board %s, got %s", tt.board, crawler.Board)
				}
				if crawler.Pages != tt.pages {
					t.Errorf("Expected pages %d, got %d", tt.pages, crawler.Pages)
				}
				if crawler.PushRate != tt.pushRate {
					t.Errorf("Expected pushRate %d, got %d", tt.pushRate, crawler.PushRate)
				}
				if crawler.FileURL != tt.fileURL {
					t.Errorf("Expected fileURL %s, got %s", tt.fileURL, crawler.FileURL)
				}
				if crawler.Client == nil {
					t.Error("Crawler should have HTTP client")
				}
				if crawler.Config == nil {
					t.Error("Crawler should have config")
				}
			}
		})
	}
}

func TestCleanFileName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"normal_filename.txt", "normal_filename.txt"},
		{"file/withslash", "filewithslash"}, // Remove backslash since regex doesn't match it
		{"file:with*special?chars", "filewithspecialchars"},
		{"file<with>pipe|chars", "filewithpipechars"},
		{"file\"with'quotes", "filewith'quotes"}, // quotes are not allowed - " is invalid
		{"[正妹] 測試標題", "[正妹] 測試標題"},               // Chinese characters should be preserved
		{"", ""},
		{"///", ""},
		{"file with spaces", "file with spaces"}, // spaces should be preserved
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := cleanFileName(tt.input)
			if result != tt.expected {
				t.Errorf("cleanFileName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestInvalidCharsRegex(t *testing.T) {
	// Test the regex pattern directly
	invalidChars := regexp.MustCompile(`[\/:*?"<>|]`)

	tests := []struct {
		char     string
		expected bool
	}{
		{"/", true},
		{"\\", false}, // backslash is not in the regex pattern
		{":", true},
		{"*", true},
		{"?", true},
		{"\"", true},
		{"<", true},
		{">", true},
		{"|", true},
		{"a", false},
		{"1", false},
		{" ", false},
		{"_", false},
		{"-", false},
		{".", false},
		{"[", false},
		{"]", false},
		{"(", false},
		{")", false},
	}

	for _, tt := range tests {
		t.Run(tt.char, func(t *testing.T) {
			result := invalidChars.MatchString(tt.char)
			if result != tt.expected {
				t.Errorf("invalidChars.MatchString(%q) = %v, want %v", tt.char, result, tt.expected)
			}
		})
	}
}

func TestCrawlerWithContext(t *testing.T) {
	cfg := config.DefaultConfig()

	// Create a minimal config for faster testing
	cfg.Crawler.Workers = 1
	cfg.Crawler.ParserCount = 1
	cfg.Crawler.Delays.MinMs = 100
	cfg.Crawler.Delays.MaxMs = 200

	crawler, err := NewCrawler("Beauty", 1, 10, "", cfg)
	if err != nil {
		t.Fatalf("Failed to create crawler: %v", err)
	}

	// Test context cancellation
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Run crawler - it should stop quickly due to context timeout
	start := time.Now()
	crawler.Run(ctx)
	elapsed := time.Since(start)

	// Should finish quickly due to context cancellation
	if elapsed > 2*time.Second {
		t.Errorf("Crawler took too long to stop: %v", elapsed)
	}
}

func TestCrawlerFileMode(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Crawler.Workers = 1
	cfg.Crawler.ParserCount = 1

	// Create a temporary file with test URLs
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_urls.txt")

	testURLs := []string{
		"https://www.ptt.cc/bbs/Beauty/M.1234567890.A.ABC.html",
		"https://www.ptt.cc/bbs/Beauty/M.1234567891.A.DEF.html",
	}

	content := ""
	for _, url := range testURLs {
		content += url + "\n"
	}

	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	crawler, err := NewCrawler("", 0, 0, testFile, cfg)
	if err != nil {
		t.Fatalf("Failed to create crawler: %v", err)
	}

	// Test file mode initialization
	if crawler.FileURL != testFile {
		t.Errorf("Expected FileURL %s, got %s", testFile, crawler.FileURL)
	}

	// Test with short timeout to avoid long running test
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	crawler.Run(ctx)
	// Test passed if no panic occurred
}

func TestCrawlerConfigValidation(t *testing.T) {
	// Test that crawler respects configuration values
	cfg := &config.Config{
		Crawler: config.CrawlerConfig{
			Workers:     5,
			ParserCount: 3,
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
				Timeout:               "30s",
				MaxIdleConns:          100,
				MaxIdleConnsPerHost:   20,
				IdleConnTimeout:       "90s",
				TLSHandshakeTimeout:   "10s",
				ExpectContinueTimeout: "1s",
			},
		},
	}

	crawler, err := NewCrawler("Beauty", 1, 10, "", cfg)
	if err != nil {
		t.Fatalf("Failed to create crawler: %v", err)
	}

	// Verify configuration is properly set
	if crawler.Config.Crawler.Workers != 5 {
		t.Errorf("Expected workers 5, got %d", crawler.Config.Crawler.Workers)
	}
	if crawler.Config.Crawler.ParserCount != 3 {
		t.Errorf("Expected parserCount 3, got %d", crawler.Config.Crawler.ParserCount)
	}

	// Test delay range
	minDelay, maxDelay := crawler.Config.GetDelayRange()
	expectedMin := 1000 * time.Millisecond
	expectedMax := 2000 * time.Millisecond

	if minDelay != expectedMin {
		t.Errorf("Expected min delay %v, got %v", expectedMin, minDelay)
	}
	if maxDelay != expectedMax {
		t.Errorf("Expected max delay %v, got %v", expectedMax, maxDelay)
	}
}

func TestCrawlerErrorHandling(t *testing.T) {
	// Test crawler with empty file path (board mode)
	cfg := config.DefaultConfig()
	cfg.Crawler.Workers = 1
	cfg.Crawler.ParserCount = 1

	// Test board mode with short timeout - should handle network errors gracefully
	crawler, err := NewCrawler("NonExistentBoard", 1, 10, "", cfg)
	if err != nil {
		t.Fatalf("Failed to create crawler: %v", err)
	}

	// Should handle network errors gracefully
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Crawler panicked: %v", r)
		}
	}()

	crawler.Run(ctx)
	// Test passes if no panic occurred
}

func TestCrawlerChannelBuffers(t *testing.T) {
	// Test that channels are created with correct buffer sizes
	cfg := &config.Config{
		Crawler: config.CrawlerConfig{
			Workers:     2,
			ParserCount: 2,
			Channels: config.ChannelConfig{
				ArticleInfo:  10,
				DownloadTask: 20,
				MarkdownTask: 5,
			},
			Delays: config.DelayConfig{
				MinMs: 100,
				MaxMs: 200,
			},
			HTTP: config.HTTPConfig{
				Timeout:               "5s",
				MaxIdleConns:          50,
				MaxIdleConnsPerHost:   10,
				IdleConnTimeout:       "30s",
				TLSHandshakeTimeout:   "5s",
				ExpectContinueTimeout: "1s",
			},
		},
	}

	crawler, err := NewCrawler("Beauty", 1, 10, "", cfg)
	if err != nil {
		t.Fatalf("Failed to create crawler: %v", err)
	}

	// Test that crawler can be created and configured properly
	if crawler.Config.Crawler.Channels.ArticleInfo != 10 {
		t.Errorf("Expected ArticleInfo channel buffer 10, got %d", crawler.Config.Crawler.Channels.ArticleInfo)
	}
	if crawler.Config.Crawler.Channels.DownloadTask != 20 {
		t.Errorf("Expected DownloadTask channel buffer 20, got %d", crawler.Config.Crawler.Channels.DownloadTask)
	}
	if crawler.Config.Crawler.Channels.MarkdownTask != 5 {
		t.Errorf("Expected MarkdownTask channel buffer 5, got %d", crawler.Config.Crawler.Channels.MarkdownTask)
	}
}
