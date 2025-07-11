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
			name:       "Load invalid config returns default",
			configPath: "../tests/fixtures/config_invalid.yaml",
			validate: func(t *testing.T, cfg *Config) {
				// Should return default config when invalid
				if cfg.Crawler.Workers != 10 {
					t.Errorf("expected default workers = 10, got %d", cfg.Crawler.Workers)
				}
			},
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
			// Load function never returns error based on implementation
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
			min, max := tt.config.GetDelayRange()
			if min != tt.expectMin {
				t.Errorf("GetDelayRange() min = %v, want %v", min, tt.expectMin)
			}
			if max != tt.expectMax {
				t.Errorf("GetDelayRange() max = %v, want %v", max, tt.expectMax)
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
