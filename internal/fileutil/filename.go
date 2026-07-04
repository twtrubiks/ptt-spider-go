// Package fileutil 提供從圖片 URL 推導本地檔名的共用邏輯，
// 供 crawler（下載存檔）與 markdown（產生圖片連結）使用，
// 確保下載的檔名與 README 中的連結一致。
package fileutil

import (
	"net/url"
	"path"
	"strings"
)

// ImageFileName 從圖片 URL 推導本地儲存檔名。
// 以 URL path 的最後一段為檔名（忽略 query string 與 fragment），
// imgur 無副檔名的連結會補上 .jpg。
func ImageFileName(imgURL string) string {
	name := path.Base(imgURL)
	if u, err := url.Parse(imgURL); err == nil {
		name = path.Base(u.Path)
	}
	if strings.Contains(imgURL, "imgur.com") && !strings.Contains(name, ".") {
		name += ".jpg"
	}
	return name
}
