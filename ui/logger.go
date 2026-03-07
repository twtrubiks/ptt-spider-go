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
