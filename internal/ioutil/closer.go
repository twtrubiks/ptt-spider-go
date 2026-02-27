// Package ioutil 提供共用的 I/O 工具函式。
package ioutil

import (
	"io"
	"log"
)

// CloseWithLog 關閉資源並記錄錯誤。
func CloseWithLog(closer io.Closer, name string) {
	if err := closer.Close(); err != nil {
		log.Printf("關閉 %s 失敗: %v", name, err)
	}
}
