// Package fileutil 提供從圖片 URL 推導本地檔名的共用邏輯，
// 供 crawler（下載存檔）與 markdown（產生圖片連結）使用，
// 確保下載的檔名與 README 中的連結一致。
package fileutil

import (
	"fmt"
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

// ImageFileNames 將圖片 URL 列表轉換為本地檔名列表，與輸入一一對應。
// 不同 URL 推導出相同檔名時，後者在副檔名前加上 _2、_3… 序號後綴，
// 避免同一目錄下互相覆蓋。給定相同輸入時輸出為確定性結果，
// crawler 與 markdown 以同一列表呼叫即可得到一致的檔名。
func ImageFileNames(imgURLs []string) []string {
	names := make([]string, 0, len(imgURLs))
	taken := make(map[string]struct{}, len(imgURLs))
	for _, imgURL := range imgURLs {
		base := ImageFileName(imgURL)
		ext := path.Ext(base)
		stem := strings.TrimSuffix(base, ext)
		name := base
		for i := 2; ; i++ {
			if _, ok := taken[name]; !ok {
				break
			}
			name = fmt.Sprintf("%s_%d%s", stem, i, ext)
		}
		taken[name] = struct{}{}
		names = append(names, name)
	}
	return names
}
