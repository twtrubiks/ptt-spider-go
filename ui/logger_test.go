package ui

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

func TestPlainLogger_ImplementsInterface(_ *testing.T) {
	var _ Logger = (*PlainLogger)(nil)
	var _ Logger = NewPlainLogger()
}

func TestStyledLogger_ImplementsInterface(_ *testing.T) {
	var _ Logger = (*StyledLogger)(nil)
	var _ Logger = NewStyledLogger()
}

func TestPlainLogger_Output(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(nil)

	logger := NewPlainLogger()

	logger.Info("test info %s", "msg")
	logger.Success("test success %d", 42)
	logger.Error("test error")
	logger.Warn("test warn %s %s", "a", "b")

	output := buf.String()

	tests := []struct {
		name     string
		expected string
	}{
		{"info message", "test info msg"},
		{"success message", "test success 42"},
		{"error message", "test error"},
		{"warn message", "test warn a b"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(output, tt.expected) {
				t.Errorf("expected output to contain %q, got %q", tt.expected, output)
			}
		})
	}
}

func TestStyledLogger_NoPanic(t *testing.T) {
	logger := NewStyledLogger()

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("StyledLogger panicked: %v", r)
		}
	}()

	logger.Info("test %s", "info")
	logger.Success("test %d", 1)
	logger.Error("test error")
	logger.Warn("test %s %s", "a", "b")
}

func TestStyledLogger_NoFormatArgs(t *testing.T) {
	logger := NewStyledLogger()

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("StyledLogger panicked with no format args: %v", r)
		}
	}()

	logger.Info("simple message")
	logger.Success("done")
	logger.Error("failed")
	logger.Warn("warning")
}
