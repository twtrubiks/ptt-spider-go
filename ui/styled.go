package ui

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	timeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	infoLabel    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	successLabel = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
	errorLabel   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("9"))
	warnLabel    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11"))

	infoText    = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	successText = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	errorText   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	warnText    = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
)

// StyledLogger 使用 Lip Gloss 樣式化輸出。
// 適用於互動式終端環境，提供彩色分級日誌。
// 內建 mutex 確保多 goroutine 並發寫入不會產生交錯輸出。
type StyledLogger struct {
	mu sync.Mutex
}

// NewStyledLogger 建立帶 Lip Gloss 樣式的 Logger.
func NewStyledLogger() *StyledLogger {
	return &StyledLogger{}
}

// Info 輸出藍色一般資訊訊息.
func (l *StyledLogger) Info(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	l.mu.Lock()
	fmt.Fprintf(os.Stderr, "%s %s  %s\n", timestamp(), infoLabel.Render("INFO"), infoText.Render(msg))
	l.mu.Unlock()
}

// Success 輸出綠色成功訊息.
func (l *StyledLogger) Success(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	l.mu.Lock()
	fmt.Fprintf(os.Stderr, "%s %s  %s\n", timestamp(), successLabel.Render(" OK "), successText.Render(msg))
	l.mu.Unlock()
}

// Error 輸出紅色錯誤訊息.
func (l *StyledLogger) Error(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	l.mu.Lock()
	fmt.Fprintf(os.Stderr, "%s %s  %s\n", timestamp(), errorLabel.Render(" ERR"), errorText.Render(msg))
	l.mu.Unlock()
}

// Warn 輸出黃色警告訊息.
func (l *StyledLogger) Warn(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	l.mu.Lock()
	fmt.Fprintf(os.Stderr, "%s %s  %s\n", timestamp(), warnLabel.Render("WARN"), warnText.Render(msg))
	l.mu.Unlock()
}

func timestamp() string {
	return timeStyle.Render(time.Now().Format("15:04:05"))
}
