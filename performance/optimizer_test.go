package performance

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// testLogger 測試用的 noop logger
type testLogger struct{}

func (testLogger) Info(string, ...any) {}

func TestNewOptimizer(t *testing.T) {
	opt := NewOptimizer(5*time.Second, testLogger{})
	if opt == nil {
		t.Fatal("NewOptimizer returned nil")
	}
	if opt.monitorInterval != 5*time.Second {
		t.Errorf("monitorInterval = %v, want %v", opt.monitorInterval, 5*time.Second)
	}
}

func TestOptimizer_StartAndStop(t *testing.T) {
	opt := NewOptimizer(50*time.Millisecond, testLogger{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	opt.Start(ctx)

	// 等待至少一個 tick 確認 ticker 在 goroutine 內正常運作
	time.Sleep(120 * time.Millisecond)

	// Stop 不應 panic
	opt.Stop()
}

func TestOptimizer_StopMultipleCalls(t *testing.T) {
	opt := NewOptimizer(time.Second, testLogger{})
	ctx := context.Background()
	opt.Start(ctx)

	// 多次呼叫 Stop 不應 panic
	opt.Stop()
	opt.Stop()
	opt.Stop()
}

func TestOptimizer_StartContextCancel(t *testing.T) {
	opt := NewOptimizer(50*time.Millisecond, testLogger{})
	ctx, cancel := context.WithCancel(context.Background())

	opt.Start(ctx)
	time.Sleep(80 * time.Millisecond)

	// 取消 context 應讓 goroutine 結束
	cancel()
	time.Sleep(80 * time.Millisecond)
}

func TestOptimizer_MonitorLogsMemory(t *testing.T) {
	opt := NewOptimizer(30*time.Millisecond, testLogger{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	opt.Start(ctx)

	// 等待足夠時間讓 ticker 至少觸發幾次
	time.Sleep(150 * time.Millisecond)
	opt.Stop()

	// 透過 GetMemoryStats 確認可正常取得資訊
	stats := opt.GetMemoryStats()
	if stats.Alloc == 0 {
		t.Error("expected non-zero memory allocation")
	}
}

func TestGetMemoryStats(t *testing.T) {
	opt := NewOptimizer(time.Second, testLogger{})
	stats := opt.GetMemoryStats()

	if stats.Alloc == 0 {
		t.Error("expected non-zero Alloc")
	}
	if stats.Sys == 0 {
		t.Error("expected non-zero Sys")
	}
	if stats.NumGoroutine == 0 {
		t.Error("expected non-zero NumGoroutine")
	}
}

func TestMemoryStats_String(t *testing.T) {
	stats := MemoryStats{
		Alloc:        1024 * 1024,
		Sys:          2 * 1024 * 1024,
		NumGC:        5,
		NumGoroutine: 10,
	}

	str := stats.String()
	if str == "" {
		t.Error("expected non-empty string")
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input    uint64
		expected string
	}{
		{0, "0 B"},
		{500, "500 B"},
		{1024, "1.0 KiB"},
		{1024 * 1024, "1.0 MiB"},
		{1024 * 1024 * 1024, "1.0 GiB"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d", tt.input), func(t *testing.T) {
			result := formatBytes(tt.input)
			if result != tt.expected {
				t.Errorf("formatBytes(%d) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
