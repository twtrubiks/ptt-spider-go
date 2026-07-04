package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name       string
		configPath string
		wantErr    bool
		validate   func(t *testing.T, cfg *Config)
	}{
		{
			name:       "Load valid config",
			configPath: "../tests/fixtures/config_valid.yaml",
			validate: func(t *testing.T, cfg *Config) {
				if cfg.Crawler.Workers != 5 {
					t.Errorf("expected workers = 5, got %d", cfg.Crawler.Workers)
				}
				if cfg.Crawler.Delays.MinMs != 1000 {
					t.Errorf("expected minMs = 1000, got %d", cfg.Crawler.Delays.MinMs)
				}
				if cfg.Crawler.HTTP.MaxIdleConns != 50 {
					t.Errorf("expected maxIdleConns = 50, got %d", cfg.Crawler.HTTP.MaxIdleConns)
				}
			},
		},
		{
			name:       "Load HTTP config",
			configPath: "../tests/fixtures/config_http_test.yaml",
			validate: func(t *testing.T, cfg *Config) {
				if cfg.Crawler.HTTP.Timeout != "45s" {
					t.Errorf("expected timeout = 45s, got %s", cfg.Crawler.HTTP.Timeout)
				}
				if cfg.Crawler.HTTP.MaxIdleConns != 200 {
					t.Errorf("expected maxIdleConns = 200, got %d", cfg.Crawler.HTTP.MaxIdleConns)
				}
				if cfg.Crawler.HTTP.MaxIdleConnsPerHost != 30 {
					t.Errorf("expected maxIdleConnsPerHost = 30, got %d", cfg.Crawler.HTTP.MaxIdleConnsPerHost)
				}
			},
		},
		{
			name:       "Load invalid config returns error",
			configPath: "../tests/fixtures/config_invalid.yaml",
			wantErr:    true,
			validate:   nil,
		},
		{
			name:       "Load non-existent config returns default",
			configPath: "../tests/fixtures/non_existent.yaml",
			validate: func(t *testing.T, cfg *Config) {
				// Check default values
				if cfg.Crawler.Workers != 10 {
					t.Errorf("expected default workers = 10, got %d", cfg.Crawler.Workers)
				}
				if cfg.Crawler.ParserCount != 10 {
					t.Errorf("expected default parserCount = 10, got %d", cfg.Crawler.ParserCount)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := Load(tt.configPath)
			if tt.wantErr {
				if err == nil {
					t.Error("Load() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("Load() unexpected error = %v", err)
				return
			}
			if cfg == nil {
				t.Fatal("Load() returned nil config")
			}
			if tt.validate != nil {
				tt.validate(t, cfg)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Test crawler defaults
	if cfg.Crawler.Workers != 10 {
		t.Errorf("expected default workers = 10, got %d", cfg.Crawler.Workers)
	}
	if cfg.Crawler.ParserCount != 10 {
		t.Errorf("expected default parserCount = 10, got %d", cfg.Crawler.ParserCount)
	}

	// Test channel defaults
	if cfg.Crawler.Channels.ArticleInfo != 100 {
		t.Errorf("expected default articleInfo channel = 100, got %d", cfg.Crawler.Channels.ArticleInfo)
	}
	if cfg.Crawler.Channels.DownloadTask != 200 {
		t.Errorf("expected default downloadTask channel = 200, got %d", cfg.Crawler.Channels.DownloadTask)
	}

	// Test delay defaults
	if cfg.Crawler.Delays.MinMs != 500 {
		t.Errorf("expected default minMs = 500, got %d", cfg.Crawler.Delays.MinMs)
	}
	if cfg.Crawler.Delays.MaxMs != 2000 {
		t.Errorf("expected default maxMs = 2000, got %d", cfg.Crawler.Delays.MaxMs)
	}

	// Test HTTP defaults
	if cfg.Crawler.HTTP.Timeout != "30s" {
		t.Errorf("expected default timeout = 30s, got %s", cfg.Crawler.HTTP.Timeout)
	}
	if cfg.Crawler.HTTP.MaxIdleConns != 100 {
		t.Errorf("expected default maxIdleConns = 100, got %d", cfg.Crawler.HTTP.MaxIdleConns)
	}
}

func TestGetTimeoutDuration(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected time.Duration
	}{
		{
			name: "Valid timeout",
			config: &Config{
				Crawler: CrawlerConfig{
					HTTP: HTTPConfig{
						Timeout: "45s",
					},
				},
			},
			expected: 45 * time.Second,
		},
		{
			name: "Invalid timeout format",
			config: &Config{
				Crawler: CrawlerConfig{
					HTTP: HTTPConfig{
						Timeout: "invalid",
					},
				},
			},
			expected: 30 * time.Second, // default
		},
		{
			name: "Empty timeout",
			config: &Config{
				Crawler: CrawlerConfig{
					HTTP: HTTPConfig{
						Timeout: "",
					},
				},
			},
			expected: 30 * time.Second, // default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetTimeoutDuration()
			if result != tt.expected {
				t.Errorf("GetTimeoutDuration() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetDelayRange(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		expectMin time.Duration
		expectMax time.Duration
	}{
		{
			name: "Normal delay range",
			config: &Config{
				Crawler: CrawlerConfig{
					Delays: DelayConfig{
						MinMs: 1000,
						MaxMs: 3000,
					},
				},
			},
			expectMin: 1000 * time.Millisecond,
			expectMax: 3000 * time.Millisecond,
		},
		{
			name: "Zero delays",
			config: &Config{
				Crawler: CrawlerConfig{
					Delays: DelayConfig{
						MinMs: 0,
						MaxMs: 0,
					},
				},
			},
			expectMin: 0,
			expectMax: 0,
		},
		{
			name: "High delays",
			config: &Config{
				Crawler: CrawlerConfig{
					Delays: DelayConfig{
						MinMs: 5000,
						MaxMs: 10000,
					},
				},
			},
			expectMin: 5000 * time.Millisecond,
			expectMax: 10000 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMin, gotMax := tt.config.GetDelayRange()
			if gotMin != tt.expectMin {
				t.Errorf("GetDelayRange() min = %v, want %v", gotMin, tt.expectMin)
			}
			if gotMax != tt.expectMax {
				t.Errorf("GetDelayRange() max = %v, want %v", gotMax, tt.expectMax)
			}
		})
	}
}

func TestHTTPTimeoutParsing(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		testFunc func(*Config) time.Duration
		expected time.Duration
	}{
		{
			name: "GetIdleConnTimeout valid",
			config: &Config{
				Crawler: CrawlerConfig{
					HTTP: HTTPConfig{
						IdleConnTimeout: "120s",
					},
				},
			},
			testFunc: (*Config).GetIdleConnTimeout,
			expected: 120 * time.Second,
		},
		{
			name: "GetIdleConnTimeout invalid",
			config: &Config{
				Crawler: CrawlerConfig{
					HTTP: HTTPConfig{
						IdleConnTimeout: "invalid",
					},
				},
			},
			testFunc: (*Config).GetIdleConnTimeout,
			expected: 90 * time.Second, // default
		},
		{
			name: "GetTLSHandshakeTimeout valid",
			config: &Config{
				Crawler: CrawlerConfig{
					HTTP: HTTPConfig{
						TLSHandshakeTimeout: "15s",
					},
				},
			},
			testFunc: (*Config).GetTLSHandshakeTimeout,
			expected: 15 * time.Second,
		},
		{
			name: "GetTLSHandshakeTimeout invalid",
			config: &Config{
				Crawler: CrawlerConfig{
					HTTP: HTTPConfig{
						TLSHandshakeTimeout: "",
					},
				},
			},
			testFunc: (*Config).GetTLSHandshakeTimeout,
			expected: 10 * time.Second, // default
		},
		{
			name: "GetExpectContinueTimeout valid",
			config: &Config{
				Crawler: CrawlerConfig{
					HTTP: HTTPConfig{
						ExpectContinueTimeout: "2s",
					},
				},
			},
			testFunc: (*Config).GetExpectContinueTimeout,
			expected: 2 * time.Second,
		},
		{
			name: "GetExpectContinueTimeout empty",
			config: &Config{
				Crawler: CrawlerConfig{
					HTTP: HTTPConfig{
						ExpectContinueTimeout: "",
					},
				},
			},
			testFunc: (*Config).GetExpectContinueTimeout,
			expected: 1 * time.Second, // default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.testFunc(tt.config)
			if result != tt.expected {
				t.Errorf("%s() = %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}

func TestConfigIntegration(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_config.yaml")

	configContent := `
crawler:
  workers: 15
  parserCount: 15
  channels:
    articleInfo: 150
    downloadTask: 300
    markdownTask: 150
  delays:
    minMs: 2000
    maxMs: 4000
  http:
    timeout: "60s"
    maxIdleConns: 150
    maxIdleConnsPerHost: 25
    idleConnTimeout: "180s"
    tlsHandshakeTimeout: "20s"
    expectContinueTimeout: "3s"
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Load and test the config
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify all values
	if cfg.Crawler.Workers != 15 {
		t.Errorf("expected workers = 15, got %d", cfg.Crawler.Workers)
	}
	if cfg.Crawler.HTTP.MaxIdleConns != 150 {
		t.Errorf("expected maxIdleConns = 150, got %d", cfg.Crawler.HTTP.MaxIdleConns)
	}

	// Test timeout parsing
	if cfg.GetTimeoutDuration() != 60*time.Second {
		t.Errorf("expected timeout = 60s, got %v", cfg.GetTimeoutDuration())
	}
	if cfg.GetIdleConnTimeout() != 180*time.Second {
		t.Errorf("expected idleConnTimeout = 180s, got %v", cfg.GetIdleConnTimeout())
	}
}

// TestValidateAndFix 驗證非法數值會退回預設值，
// 避免 make(chan, -1) panic、workers=0 死鎖、parserCount=0 goroutine 洩漏。
func TestValidateAndFix(t *testing.T) {
	defaults := DefaultConfig()

	tests := []struct {
		name   string
		mutate func(*Config)
		check  func(t *testing.T, cfg *Config)
	}{
		{
			name:   "workers 為 0 退回預設",
			mutate: func(c *Config) { c.Crawler.Workers = 0 },
			check: func(t *testing.T, cfg *Config) {
				if cfg.Crawler.Workers != defaults.Crawler.Workers {
					t.Errorf("workers = %d, want %d", cfg.Crawler.Workers, defaults.Crawler.Workers)
				}
			},
		},
		{
			name:   "workers 為負數退回預設",
			mutate: func(c *Config) { c.Crawler.Workers = -5 },
			check: func(t *testing.T, cfg *Config) {
				if cfg.Crawler.Workers != defaults.Crawler.Workers {
					t.Errorf("workers = %d, want %d", cfg.Crawler.Workers, defaults.Crawler.Workers)
				}
			},
		},
		{
			name:   "parserCount 為 0 退回預設",
			mutate: func(c *Config) { c.Crawler.ParserCount = 0 },
			check: func(t *testing.T, cfg *Config) {
				if cfg.Crawler.ParserCount != defaults.Crawler.ParserCount {
					t.Errorf("parserCount = %d, want %d", cfg.Crawler.ParserCount, defaults.Crawler.ParserCount)
				}
			},
		},
		{
			name:   "channel 緩衝區為負數退回預設",
			mutate: func(c *Config) { c.Crawler.Channels.DownloadTask = -1 },
			check: func(t *testing.T, cfg *Config) {
				if cfg.Crawler.Channels.DownloadTask != defaults.Crawler.Channels.DownloadTask {
					t.Errorf("downloadTask = %d, want %d",
						cfg.Crawler.Channels.DownloadTask, defaults.Crawler.Channels.DownloadTask)
				}
			},
		},
		{
			name: "延遲為負數退回預設",
			mutate: func(c *Config) {
				c.Crawler.Delays.MinMs = -100
				c.Crawler.Delays.MaxMs = -200
			},
			check: func(t *testing.T, cfg *Config) {
				if cfg.Crawler.Delays.MinMs != defaults.Crawler.Delays.MinMs {
					t.Errorf("minMs = %d, want %d", cfg.Crawler.Delays.MinMs, defaults.Crawler.Delays.MinMs)
				}
				if cfg.Crawler.Delays.MaxMs != defaults.Crawler.Delays.MaxMs {
					t.Errorf("maxMs = %d, want %d", cfg.Crawler.Delays.MaxMs, defaults.Crawler.Delays.MaxMs)
				}
			},
		},
		{
			name:   "合法值不被修改",
			mutate: func(c *Config) { c.Crawler.Workers = 3 },
			check: func(t *testing.T, cfg *Config) {
				if cfg.Crawler.Workers != 3 {
					t.Errorf("workers = %d, want 3", cfg.Crawler.Workers)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			tt.mutate(cfg)
			cfg.validateAndFix()
			tt.check(t, cfg)
		})
	}
}

// TestLoad_InvalidNumericValues 驗證 Load 會修正配置檔中的非法數值。
func TestLoad_InvalidNumericValues(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "bad_values.yaml")

	configContent := `
crawler:
  workers: 0
  parserCount: -3
  channels:
    articleInfo: -1
    downloadTask: -1
    markdownTask: -1
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("建立測試配置檔失敗: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() unexpected error = %v", err)
	}

	defaults := DefaultConfig()
	if cfg.Crawler.Workers != defaults.Crawler.Workers {
		t.Errorf("workers = %d, want %d", cfg.Crawler.Workers, defaults.Crawler.Workers)
	}
	if cfg.Crawler.ParserCount != defaults.Crawler.ParserCount {
		t.Errorf("parserCount = %d, want %d", cfg.Crawler.ParserCount, defaults.Crawler.ParserCount)
	}
	if cfg.Crawler.Channels.ArticleInfo != defaults.Crawler.Channels.ArticleInfo {
		t.Errorf("articleInfo = %d, want %d", cfg.Crawler.Channels.ArticleInfo, defaults.Crawler.Channels.ArticleInfo)
	}
	if cfg.Crawler.Channels.DownloadTask != defaults.Crawler.Channels.DownloadTask {
		t.Errorf("downloadTask = %d, want %d", cfg.Crawler.Channels.DownloadTask, defaults.Crawler.Channels.DownloadTask)
	}
	if cfg.Crawler.Channels.MarkdownTask != defaults.Crawler.Channels.MarkdownTask {
		t.Errorf("markdownTask = %d, want %d", cfg.Crawler.Channels.MarkdownTask, defaults.Crawler.Channels.MarkdownTask)
	}
}
