package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestErrorType_String(t *testing.T) {
	tests := []struct {
		name      string
		errorType ErrorType
		want      string
	}{
		{"NetworkError", NetworkError, "NetworkError"},
		{"ParseError", ParseError, "ParseError"},
		{"FileError", FileError, "FileError"},
		{"ConfigError", ConfigError, "ConfigError"},
		{"ValidationError", ValidationError, "ValidationError"},
		{"UnknownError", ErrorType(999), "UnknownError"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.errorType.String(); got != tt.want {
				t.Errorf("ErrorType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCrawlerError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *CrawlerError
		want string
	}{
		{
			name: "error without cause",
			err: &CrawlerError{
				Type:    NetworkError,
				Message: "connection failed",
			},
			want: "[NetworkError] connection failed",
		},
		{
			name: "error with cause",
			err: &CrawlerError{
				Type:    ParseError,
				Message: "parsing failed",
				Cause:   fmt.Errorf("invalid syntax"),
			},
			want: "[ParseError] parsing failed: invalid syntax",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("CrawlerError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCrawlerError_Unwrap(t *testing.T) {
	cause := fmt.Errorf("root cause")
	err := &CrawlerError{
		Type:    NetworkError,
		Message: "test error",
		Cause:   cause,
	}

	if got := err.Unwrap(); got != cause {
		t.Errorf("CrawlerError.Unwrap() = %v, want %v", got, cause)
	}

	errNoCause := &CrawlerError{
		Type:    NetworkError,
		Message: "test error",
	}

	if got := errNoCause.Unwrap(); got != nil {
		t.Errorf("CrawlerError.Unwrap() = %v, want nil", got)
	}
}

func TestCrawlerError_WithContext(t *testing.T) {
	err := &CrawlerError{
		Type:    NetworkError,
		Message: "test error",
	}

	result := err.WithContext("url", "https://example.com")

	if result != err {
		t.Error("WithContext should return the same error instance")
	}

	if err.Context == nil {
		t.Error("Context should be initialized")
	}

	value, exists := err.GetContext("url")
	if !exists {
		t.Error("Context value should exist")
	}

	if value != "https://example.com" {
		t.Errorf("Context value = %v, want 'https://example.com'", value)
	}
}

func TestCrawlerError_GetContext(t *testing.T) {
	err := &CrawlerError{
		Type:    NetworkError,
		Message: "test error",
		Context: map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		},
	}

	// Test existing key
	value, exists := err.GetContext("key1")
	if !exists {
		t.Error("key1 should exist")
	}
	if value != "value1" {
		t.Errorf("Context value = %v, want 'value1'", value)
	}

	// Test non-existing key
	_, exists = err.GetContext("nonexistent")
	if exists {
		t.Error("nonexistent key should not exist")
	}

	// Test nil context
	errNilContext := &CrawlerError{}
	_, exists = errNilContext.GetContext("key")
	if exists {
		t.Error("key should not exist in nil context")
	}
}

func TestCrawlerError_Is(t *testing.T) {
	networkErr := &CrawlerError{Type: NetworkError}
	parseErr := &CrawlerError{Type: ParseError}
	regularErr := fmt.Errorf("regular error")

	if !networkErr.Is(networkErr) {
		t.Error("error should match itself")
	}

	if !networkErr.Is(&CrawlerError{Type: NetworkError}) {
		t.Error("error should match same type")
	}

	if networkErr.Is(parseErr) {
		t.Error("error should not match different type")
	}

	if networkErr.Is(regularErr) {
		t.Error("error should not match regular error")
	}
}

func TestNewCrawlerError(t *testing.T) {
	err := NewCrawlerError(NetworkError, "test message")

	if err.Type != NetworkError {
		t.Errorf("Type = %v, want %v", err.Type, NetworkError)
	}

	if err.Message != "test message" {
		t.Errorf("Message = %v, want 'test message'", err.Message)
	}

	if err.Cause != nil {
		t.Errorf("Cause = %v, want nil", err.Cause)
	}
}

func TestNewCrawlerErrorWithCause(t *testing.T) {
	cause := fmt.Errorf("root cause")
	err := NewCrawlerErrorWithCause(ParseError, "test message", cause)

	if err.Type != ParseError {
		t.Errorf("Type = %v, want %v", err.Type, ParseError)
	}

	if err.Message != "test message" {
		t.Errorf("Message = %v, want 'test message'", err.Message)
	}

	if err.Cause != cause {
		t.Errorf("Cause = %v, want %v", err.Cause, cause)
	}
}

func TestConvenienceFunctions(t *testing.T) {
	cause := fmt.Errorf("test cause")

	// Test NewNetworkError
	netErr := NewNetworkError("network failed", cause)
	if netErr.Type != NetworkError {
		t.Errorf("NewNetworkError type = %v, want %v", netErr.Type, NetworkError)
	}

	// Test NewParseError
	parseErr := NewParseError("parse failed", cause)
	if parseErr.Type != ParseError {
		t.Errorf("NewParseError type = %v, want %v", parseErr.Type, ParseError)
	}

	// Test NewFileError
	fileErr := NewFileError("file failed", cause)
	if fileErr.Type != FileError {
		t.Errorf("NewFileError type = %v, want %v", fileErr.Type, FileError)
	}

	// Test NewConfigError
	configErr := NewConfigError("config failed", cause)
	if configErr.Type != ConfigError {
		t.Errorf("NewConfigError type = %v, want %v", configErr.Type, ConfigError)
	}

	// Test NewValidationError
	validErr := NewValidationError("validation failed")
	if validErr.Type != ValidationError {
		t.Errorf("NewValidationError type = %v, want %v", validErr.Type, ValidationError)
	}
	if validErr.Cause != nil {
		t.Errorf("NewValidationError cause = %v, want nil", validErr.Cause)
	}
}

func TestErrorTypeCheckers(t *testing.T) {
	networkErr := NewNetworkError("network error", nil)
	parseErr := NewParseError("parse error", nil)
	fileErr := NewFileError("file error", nil)
	configErr := NewConfigError("config error", nil)
	validErr := NewValidationError("validation error")
	regularErr := fmt.Errorf("regular error")

	// Test IsNetworkError
	if !IsNetworkError(networkErr) {
		t.Error("IsNetworkError should return true for network error")
	}
	if IsNetworkError(parseErr) {
		t.Error("IsNetworkError should return false for non-network error")
	}
	if IsNetworkError(regularErr) {
		t.Error("IsNetworkError should return false for regular error")
	}

	// Test IsParseError
	if !IsParseError(parseErr) {
		t.Error("IsParseError should return true for parse error")
	}
	if IsParseError(networkErr) {
		t.Error("IsParseError should return false for non-parse error")
	}

	// Test IsFileError
	if !IsFileError(fileErr) {
		t.Error("IsFileError should return true for file error")
	}
	if IsFileError(networkErr) {
		t.Error("IsFileError should return false for non-file error")
	}

	// Test IsConfigError
	if !IsConfigError(configErr) {
		t.Error("IsConfigError should return true for config error")
	}
	if IsConfigError(networkErr) {
		t.Error("IsConfigError should return false for non-config error")
	}

	// Test IsValidationError
	if !IsValidationError(validErr) {
		t.Error("IsValidationError should return true for validation error")
	}
	if IsValidationError(networkErr) {
		t.Error("IsValidationError should return false for non-validation error")
	}
}

func TestErrorsCompatibility(t *testing.T) {
	cause := fmt.Errorf("root cause")
	err := NewNetworkError("network error", cause)

	// Test errors.Is compatibility
	if !errors.Is(err, err) {
		t.Error("errors.Is should work with CrawlerError")
	}

	// Test errors.Unwrap compatibility
	if unwrapped := errors.Unwrap(err); unwrapped != cause {
		t.Errorf("errors.Unwrap() = %v, want %v", unwrapped, cause)
	}
}
