// Package ui 提供日誌輸出的抽象層，支援純文字和彩色樣式化兩種模式。
package ui

import "log"

// Logger 定義日誌輸出介面，支援不同等級的訊息樣式化。
// 實作此介面可自訂輸出格式（如彩色終端、TUI 等）。
type Logger interface {
	Info(format string, args ...any)
	Success(format string, args ...any)
	Error(format string, args ...any)
	Warn(format string, args ...any)
}

// NoopLogger 靜默丟棄所有輸出，適用於 TUI 即時進度模式。
// 在 Bubble Tea 接管 terminal 時使用，避免 log 輸出干擾 TUI 畫面。
type NoopLogger struct{}

// NewNoopLogger 建立靜默 Logger.
func NewNoopLogger() *NoopLogger { return &NoopLogger{} }

// Info 靜默丟棄 Info 訊息.
func (l *NoopLogger) Info(_ string, _ ...any) {}

// Success 靜默丟棄 Success 訊息.
func (l *NoopLogger) Success(_ string, _ ...any) {}

// Error 靜默丟棄 Error 訊息.
func (l *NoopLogger) Error(_ string, _ ...any) {}

// Warn 靜默丟棄 Warn 訊息.
func (l *NoopLogger) Warn(_ string, _ ...any) {}

// PlainLogger 使用標準 log 套件輸出，不帶任何樣式。
// 適用於自動化腳本、測試和非互動環境。
type PlainLogger struct{}

// NewPlainLogger 建立純文字 Logger.
func NewPlainLogger() *PlainLogger {
	return &PlainLogger{}
}

// Info 輸出一般資訊訊息.
func (l *PlainLogger) Info(format string, args ...any) {
	log.Printf(format, args...)
}

// Success 輸出成功訊息.
func (l *PlainLogger) Success(format string, args ...any) {
	log.Printf(format, args...)
}

// Error 輸出錯誤訊息.
func (l *PlainLogger) Error(format string, args ...any) {
	log.Printf(format, args...)
}

// Warn 輸出警告訊息.
func (l *PlainLogger) Warn(format string, args ...any) {
	log.Printf(format, args...)
}
