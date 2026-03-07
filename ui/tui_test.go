package ui

import "testing"

func TestValidateNonEmpty(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"empty is invalid", "", true},
		{"non-empty is valid", "urls.txt", false},
		{"whitespace is valid", " ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateNonEmpty(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateNonEmpty(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidatePositiveInt(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"empty uses default", "", false},
		{"valid positive", "5", false},
		{"valid large", "100", false},
		{"one", "1", false},
		{"zero is invalid", "0", true},
		{"negative", "-1", true},
		{"not a number", "abc", true},
		{"float", "3.5", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePositiveInt(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePositiveInt(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateNonNegativeInt(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"empty uses default", "", false},
		{"valid positive", "5", false},
		{"zero is valid", "0", false},
		{"negative", "-1", true},
		{"not a number", "abc", true},
		{"float", "3.5", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateNonNegativeInt(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateNonNegativeInt(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestStartupConfig_ZeroValue(t *testing.T) {
	cfg := &StartupConfig{}

	if cfg.Board != "" {
		t.Errorf("Expected empty Board, got %q", cfg.Board)
	}
	if cfg.Pages != 0 {
		t.Errorf("Expected 0 Pages, got %d", cfg.Pages)
	}
	if cfg.PushRate != 0 {
		t.Errorf("Expected 0 PushRate, got %d", cfg.PushRate)
	}
	if cfg.FileURL != "" {
		t.Errorf("Expected empty FileURL, got %q", cfg.FileURL)
	}
}

func TestApplyDefaults_Board(t *testing.T) {
	tests := []struct {
		name         string
		board        string
		pagesStr     string
		pushRateStr  string
		defaultBoard string
		defaultPages int
		defaultPush  int
		wantBoard    string
		wantPages    int
		wantPush     int
	}{
		{
			name:         "all empty uses defaults",
			defaultBoard: "beauty", defaultPages: 3, defaultPush: 10,
			wantBoard: "beauty", wantPages: 3, wantPush: 10,
		},
		{
			name:         "custom values",
			board:        "gossiping",
			pagesStr:     "5",
			pushRateStr:  "20",
			defaultBoard: "beauty", defaultPages: 3, defaultPush: 10,
			wantBoard: "gossiping", wantPages: 5, wantPush: 20,
		},
		{
			name:         "partial custom",
			board:        "nba",
			pagesStr:     "",
			pushRateStr:  "0",
			defaultBoard: "beauty", defaultPages: 3, defaultPush: 10,
			wantBoard: "nba", wantPages: 3, wantPush: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := applyBoardDefaults(tt.board, tt.pagesStr, tt.pushRateStr, tt.defaultBoard, tt.defaultPages, tt.defaultPush)
			if cfg.Board != tt.wantBoard {
				t.Errorf("Board = %q, want %q", cfg.Board, tt.wantBoard)
			}
			if cfg.Pages != tt.wantPages {
				t.Errorf("Pages = %d, want %d", cfg.Pages, tt.wantPages)
			}
			if cfg.PushRate != tt.wantPush {
				t.Errorf("PushRate = %d, want %d", cfg.PushRate, tt.wantPush)
			}
		})
	}
}

func TestApplyDefaults_File(t *testing.T) {
	tests := []struct {
		name         string
		fileURL      string
		board        string
		defaultBoard string
		wantBoard    string
		wantFileURL  string
	}{
		{
			name:         "board uses default when empty",
			fileURL:      "urls.txt",
			defaultBoard: "beauty",
			wantBoard:    "beauty",
			wantFileURL:  "urls.txt",
		},
		{
			name:         "custom board",
			fileURL:      "urls.txt",
			board:        "gossiping",
			defaultBoard: "beauty",
			wantBoard:    "gossiping",
			wantFileURL:  "urls.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := applyFileDefaults(tt.fileURL, tt.board, tt.defaultBoard)
			if cfg.Board != tt.wantBoard {
				t.Errorf("Board = %q, want %q", cfg.Board, tt.wantBoard)
			}
			if cfg.FileURL != tt.wantFileURL {
				t.Errorf("FileURL = %q, want %q", cfg.FileURL, tt.wantFileURL)
			}
		})
	}
}
