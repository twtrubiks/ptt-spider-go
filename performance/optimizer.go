// Package performance 提供效能監控工具，
// 包括記憶體狀態監控和統計資訊。
package performance

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"
)

// Optimizer 效能監控器
type Optimizer struct {
	monitorInterval time.Duration
	stopChan        chan struct{}
	stopOnce        sync.Once
}

// NewOptimizer 建立新的效能監控器
// memoryThresholdMB 參數已棄用，保留以維持 API 相容性。
func NewOptimizer(_ int64, monitorInterval time.Duration) *Optimizer {
	return &Optimizer{
		monitorInterval: monitorInterval,
		stopChan:        make(chan struct{}),
	}
}

// Start 啟動效能監控
func (o *Optimizer) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(o.monitorInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-o.stopChan:
				return
			case <-ticker.C:
				stats := o.GetMemoryStats()
				log.Printf("記憶體監控: %s", stats.String())
			}
		}
	}()
}

// Stop 停止效能監控
func (o *Optimizer) Stop() {
	o.stopOnce.Do(func() {
		close(o.stopChan)
	})
}

// GetMemoryStats 獲取記憶體統計資訊
func (o *Optimizer) GetMemoryStats() MemoryStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return MemoryStats{
		Alloc:        m.Alloc,
		TotalAlloc:   m.TotalAlloc,
		Sys:          m.Sys,
		NumGC:        m.NumGC,
		LastGC:       time.Unix(0, int64(m.LastGC)),
		HeapAlloc:    m.HeapAlloc,
		HeapSys:      m.HeapSys,
		HeapIdle:     m.HeapIdle,
		HeapInuse:    m.HeapInuse,
		NumGoroutine: runtime.NumGoroutine(),
	}
}

// MemoryStats 記憶體統計資訊
type MemoryStats struct {
	Alloc        uint64    // 當前分配的記憶體 (bytes)
	TotalAlloc   uint64    // 總分配的記憶體 (bytes)
	Sys          uint64    // 系統記憶體 (bytes)
	NumGC        uint32    // GC 次數
	LastGC       time.Time // 最後一次 GC 時間
	HeapAlloc    uint64    // 堆記憶體分配 (bytes)
	HeapSys      uint64    // 堆系統記憶體 (bytes)
	HeapIdle     uint64    // 堆空閒記憶體 (bytes)
	HeapInuse    uint64    // 堆使用中記憶體 (bytes)
	NumGoroutine int       // Goroutine 數量
}

// String 返回記憶體統計的字串表示
func (ms MemoryStats) String() string {
	return fmt.Sprintf(
		"Memory: Alloc=%s, Sys=%s, NumGC=%d, Goroutines=%d",
		formatBytes(ms.Alloc),
		formatBytes(ms.Sys),
		ms.NumGC,
		ms.NumGoroutine,
	)
}

// formatBytes 格式化 bytes 為可讀格式
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
