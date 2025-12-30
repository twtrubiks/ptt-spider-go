// Package performance 提供效能優化工具，
// 包括記憶體監控、自動垃圾回收和 HTTP 連線池配置。
package performance

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

// Optimizer 效能優化器
type Optimizer struct {
	memoryThreshold int64 // 記憶體閾值 (MB)
	gcInterval      time.Duration
	stopChan        chan struct{}
	mu              sync.RWMutex
}

// NewOptimizer 建立新的效能優化器
func NewOptimizer(memoryThresholdMB int64, gcInterval time.Duration) *Optimizer {
	return &Optimizer{
		memoryThreshold: memoryThresholdMB * 1024 * 1024, // 轉換為 bytes
		gcInterval:      gcInterval,
		stopChan:        make(chan struct{}),
	}
}

// Start 啟動效能監控和優化
func (o *Optimizer) Start(ctx context.Context) {
	ticker := time.NewTicker(o.gcInterval)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-o.stopChan:
				return
			case <-ticker.C:
				o.optimizeMemory()
			}
		}
	}()
}

// Stop 停止效能監控
func (o *Optimizer) Stop() {
	close(o.stopChan)
}

// optimizeMemory 記憶體優化
func (o *Optimizer) optimizeMemory() {
	o.mu.Lock()
	defer o.mu.Unlock()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 如果記憶體使用超過閾值，執行 GC
	if int64(m.Alloc) > o.memoryThreshold {
		runtime.GC()
		debug.FreeOSMemory()
	}
}

// GetMemoryStats 獲取記憶體統計資訊
func (o *Optimizer) GetMemoryStats() MemoryStats {
	o.mu.RLock()
	defer o.mu.RUnlock()

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

// ConnectionPool 連線池管理
type ConnectionPool struct {
	maxIdleConns        int
	maxIdleConnsPerHost int
	idleConnTimeout     time.Duration
	tlsHandshakeTimeout time.Duration
}

// NewConnectionPool 建立新的連線池配置
func NewConnectionPool(maxIdle, maxIdlePerHost int, idleTimeout, tlsTimeout time.Duration) *ConnectionPool {
	return &ConnectionPool{
		maxIdleConns:        maxIdle,
		maxIdleConnsPerHost: maxIdlePerHost,
		idleConnTimeout:     idleTimeout,
		tlsHandshakeTimeout: tlsTimeout,
	}
}

// OptimizeTransport 優化 HTTP Transport
func (cp *ConnectionPool) OptimizeTransport(transport *http.Transport) {
	transport.MaxIdleConns = cp.maxIdleConns
	transport.MaxIdleConnsPerHost = cp.maxIdleConnsPerHost
	transport.IdleConnTimeout = cp.idleConnTimeout
	transport.TLSHandshakeTimeout = cp.tlsHandshakeTimeout
	transport.DisableKeepAlives = false
	transport.DisableCompression = false
}
