package crawler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"testing/iotest"
	"time"

	"github.com/twtrubiks/ptt-spider-go/config"
	"github.com/twtrubiks/ptt-spider-go/mocks"
	"github.com/twtrubiks/ptt-spider-go/types"
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
			wantErr:  true, // 看板模式必須指定看板名稱
		},
		{
			name:     "Path traversal board rejected",
			board:    "../../etc",
			pages:    1,
			pushRate: 0,
			fileURL:  "",
			wantErr:  true,
		},
		{
			name:     "Board with slash rejected",
			board:    "beauty/x",
			pages:    1,
			pushRate: 0,
			fileURL:  "",
			wantErr:  true,
		},
		{
			name:     "Invalid board rejected in file mode",
			board:    "..\\..\\evil",
			pages:    0,
			pushRate: 0,
			fileURL:  "test.txt",
			wantErr:  true,
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
				if crawler.board != tt.board {
					t.Errorf("Expected board %s, got %s", tt.board, crawler.board)
				}
				if crawler.pages != tt.pages {
					t.Errorf("Expected pages %d, got %d", tt.pages, crawler.pages)
				}
				if crawler.pushRate != tt.pushRate {
					t.Errorf("Expected pushRate %d, got %d", tt.pushRate, crawler.pushRate)
				}
				if crawler.fileURL != tt.fileURL {
					t.Errorf("Expected fileURL %s, got %s", tt.fileURL, crawler.fileURL)
				}
				if crawler.client == nil {
					t.Error("Crawler should have HTTP client")
				}
				if crawler.config == nil {
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
		{"file/withslash", "filewithslash"},
		{"file\\withbackslash", "filewithbackslash"}, // Windows 上反斜線是路徑分隔符，必須過濾
		{"..\\..\\evil", "....evil"},                 // 路徑穿越防護：移除反斜線後不再構成路徑
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
	invalidChars := regexp.MustCompile(`[\\/:*?"<>|]`)

	tests := []struct {
		char     string
		expected bool
	}{
		{"/", true},
		{"\\", true}, // Windows 路徑分隔符，必須過濾以防路徑穿越
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
	if crawler.fileURL != testFile {
		t.Errorf("Expected FileURL %s, got %s", testFile, crawler.fileURL)
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
	if crawler.config.Crawler.Workers != 5 {
		t.Errorf("Expected workers 5, got %d", crawler.config.Crawler.Workers)
	}
	if crawler.config.Crawler.ParserCount != 3 {
		t.Errorf("Expected parserCount 3, got %d", crawler.config.Crawler.ParserCount)
	}

	// Test delay range
	minDelay, maxDelay := crawler.config.GetDelayRange()
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
	if crawler.config.Crawler.Channels.ArticleInfo != 10 {
		t.Errorf("Expected ArticleInfo channel buffer 10, got %d", crawler.config.Crawler.Channels.ArticleInfo)
	}
	if crawler.config.Crawler.Channels.DownloadTask != 20 {
		t.Errorf("Expected DownloadTask channel buffer 20, got %d", crawler.config.Crawler.Channels.DownloadTask)
	}
	if crawler.config.Crawler.Channels.MarkdownTask != 5 {
		t.Errorf("Expected MarkdownTask channel buffer 5, got %d", crawler.config.Crawler.Channels.MarkdownTask)
	}
}

// endlessReader 產生無限長度的資料流，用於模擬超大回應
type endlessReader struct{}

func (endlessReader) Read(p []byte) (int, error) { return len(p), nil }

func newTestCrawlerForSave(t *testing.T) *Crawler {
	t.Helper()
	return NewCrawlerWithDependencies(
		mocks.NewMockHTTPClient(),
		mocks.NewMockParser(),
		mocks.NewMockMarkdownGenerator(),
		"test", 1, 0, "", config.DefaultConfig(),
	)
}

// TestSaveToFile_ExceedsSizeLimit 驗證圖片超過大小上限時不會寫爆磁碟，且半截檔會被清除。
func TestSaveToFile_ExceedsSizeLimit(t *testing.T) {
	c := newTestCrawlerForSave(t)
	savePath := filepath.Join(t.TempDir(), "big.jpg")

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(endlessReader{}),
	}

	c.saveToFile(resp, savePath, 1)

	if _, err := os.Stat(savePath); !os.IsNotExist(err) {
		t.Errorf("超過大小上限的檔案應被刪除，但仍存在: %s", savePath)
	}
}

// TestSaveToFile_PartialDownloadRemoved 驗證下載中途失敗時半截檔會被清除。
func TestSaveToFile_PartialDownloadRemoved(t *testing.T) {
	c := newTestCrawlerForSave(t)
	savePath := filepath.Join(t.TempDir(), "partial.jpg")

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body: io.NopCloser(io.MultiReader(
			strings.NewReader("partial data"),
			iotest.ErrReader(errors.New("connection reset")),
		)),
	}

	c.saveToFile(resp, savePath, 1)

	if _, err := os.Stat(savePath); !os.IsNotExist(err) {
		t.Errorf("下載中途失敗的半截檔應被刪除，但仍存在: %s", savePath)
	}
}

// TestSaveToFile_Success 驗證正常下載會完整寫入檔案。
func TestSaveToFile_Success(t *testing.T) {
	c := newTestCrawlerForSave(t)
	savePath := filepath.Join(t.TempDir(), "ok.jpg")

	content := "fake image bytes"
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(content)),
	}

	c.saveToFile(resp, savePath, 1)

	data, err := os.ReadFile(savePath)
	if err != nil {
		t.Fatalf("讀取下載檔案失敗: %v", err)
	}
	if string(data) != content {
		t.Errorf("檔案內容 = %q, want %q", string(data), content)
	}
}

// TestArticleProducer_StopsAtFirstPage 驗證 pages 大於看板實際頁數時，
// 不會請求 index0.html、index-1.html 等不存在的頁面。
func TestArticleProducer_StopsAtFirstPage(t *testing.T) {
	var requestedURLs []string
	client := &mocks.MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			requestedURLs = append(requestedURLs, req.URL.String())
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("<html></html>")),
			}, nil
		},
	}
	parser := &mocks.MockParser{
		ParseMaxPageFunc: func(_ io.Reader) (int, error) { return 2, nil },
		ParseArticlesFunc: func(_ io.Reader) ([]types.ArticleInfo, error) {
			return nil, nil
		},
	}

	c := NewCrawlerWithDependencies(
		client, parser, mocks.NewMockMarkdownGenerator(),
		"test", 5, 0, "", config.DefaultConfig(),
	)

	ch := make(chan types.ArticleInfo, 10)
	c.articleProducer(context.Background(), ch)

	for _, u := range requestedURLs {
		if strings.Contains(u, "index0.html") || strings.Contains(u, "index-") {
			t.Errorf("不應請求不存在的頁面: %s", u)
		}
	}
	// index.html（取最大頁數）+ index2.html + index1.html
	if len(requestedURLs) != 3 {
		t.Errorf("期望請求 3 次，實際 %d 次: %v", len(requestedURLs), requestedURLs)
	}
}

// TestFetchAndParseArticle_Non200 驗證文章頁回應非 200 時回傳錯誤，
// 與 fetchMaxPage 的狀態碼檢查行為一致。
func TestFetchAndParseArticle_Non200(t *testing.T) {
	client := &mocks.MockHTTPClient{
		DoFunc: func(_ *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader("not found")),
			}, nil
		},
	}

	c := NewCrawlerWithDependencies(
		client, mocks.NewMockParser(), mocks.NewMockMarkdownGenerator(),
		"test", 1, 0, "", config.DefaultConfig(),
	)

	_, _, err := c.fetchAndParseArticle(context.Background(), types.ArticleInfo{
		URL: "https://www.ptt.cc/bbs/test/M.123.A.html",
	})
	if err == nil {
		t.Error("HTTP 404 應回傳錯誤，但收到 nil")
	}
}

// TestArticleProducerFromFile_RequiresURLPrefix 驗證檔案模式只接受
// 以 PTT 文章網址開頭的行，含子字串但非開頭的行應被略過。
func TestArticleProducerFromFile_RequiresURLPrefix(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "urls.txt")
	content := strings.Join([]string{
		"https://www.ptt.cc/bbs/Beauty/M.111.A.html",
		"  https://www.ptt.cc/bbs/Beauty/M.222.A.html  ",
		"參考 https://www.ptt.cc/bbs/Beauty/M.333.A.html 這篇",
		"https://evil.example.com/?u=https://www.ptt.cc/bbs/x",
		"",
	}, "\n")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("建立測試檔失敗: %v", err)
	}

	c := NewCrawlerWithDependencies(
		mocks.NewMockHTTPClient(), mocks.NewMockParser(), mocks.NewMockMarkdownGenerator(),
		"test", 0, 0, tmpFile, config.DefaultConfig(),
	)

	ch := make(chan types.ArticleInfo, 10)
	c.articleProducerFromFile(context.Background(), ch)

	var got []string
	for a := range ch {
		got = append(got, a.URL)
	}

	want := []string{
		"https://www.ptt.cc/bbs/Beauty/M.111.A.html",
		"https://www.ptt.cc/bbs/Beauty/M.222.A.html",
	}
	if len(got) != len(want) {
		t.Fatalf("期望 %d 個 URL，實際 %d 個: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("URL[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

// TestRunWaitsForProducer 驗證 ctx 取消時 Run 會等待 producer goroutine 結束才返回。
// 若 Run 提前返回，TUI 模式會關閉 progress channel，
// 仍在執行的 producer 隨後呼叫 emit 就會對已關閉的 channel 做 send 而 panic。
func TestRunWaitsForProducer(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})

	parser := &mocks.MockParser{
		ParseMaxPageFunc: func(io.Reader) (int, error) { return 10, nil },
		ParseArticlesFunc: func(io.Reader) ([]types.ArticleInfo, error) {
			close(entered) // 通知測試：producer 已進入列表頁解析
			<-release      // 卡住，模擬解析進行中（此期間 ctx 被取消）
			return []types.ArticleInfo{}, nil
		},
	}

	c := NewCrawlerWithDependencies(
		mocks.NewMockHTTPClient(), parser, mocks.NewMockMarkdownGenerator(),
		"test", 1, 0, "", config.DefaultConfig(),
	)
	progressCh := make(chan types.ProgressEvent, 200)
	c.progress = progressCh

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runDone := make(chan struct{})
	go func() {
		c.Run(ctx)
		close(runDone)
	}()

	<-entered // producer 正在 ParseArticles 中
	cancel()  // 模擬 Ctrl+C

	// producer 仍卡在 ParseArticles，Run 不得返回
	select {
	case <-runDone:
		t.Fatal("Run 在 producer 結束前返回，TUI 模式關閉 progress channel 後會導致 emit panic")
	case <-time.After(200 * time.Millisecond):
	}

	close(release) // 讓 producer 完成解析並結束

	select {
	case <-runDone:
	case <-time.After(5 * time.Second):
		t.Fatal("Run 未在 producer 結束後返回")
	}

	// 模擬 main.go runWithTUI：Run 返回後關閉 progress channel。
	// producer 此時必已結束，不會再有 emit。
	close(progressCh)
}
