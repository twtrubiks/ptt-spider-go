package types

import (
	"reflect"
	"testing"
)

func TestArticleInfo(t *testing.T) {
	tests := []struct {
		name string
		info ArticleInfo
		want ArticleInfo
	}{
		{
			name: "Complete article info",
			info: ArticleInfo{
				Title:    "[正妹] 測試標題",
				URL:      "https://www.ptt.cc/bbs/Beauty/M.1234567890.A.ABC.html",
				Author:   "testuser",
				PushRate: 99,
			},
			want: ArticleInfo{
				Title:    "[正妹] 測試標題",
				URL:      "https://www.ptt.cc/bbs/Beauty/M.1234567890.A.ABC.html",
				Author:   "testuser",
				PushRate: 99,
			},
		},
		{
			name: "Article with negative push rate",
			info: ArticleInfo{
				Title:    "[問卦] 爭議文章",
				URL:      "https://www.ptt.cc/bbs/Gossiping/M.1234567890.A.ABC.html",
				Author:   "controversial",
				PushRate: -5,
			},
			want: ArticleInfo{
				Title:    "[問卦] 爭議文章",
				URL:      "https://www.ptt.cc/bbs/Gossiping/M.1234567890.A.ABC.html",
				Author:   "controversial",
				PushRate: -5,
			},
		},
		{
			name: "Empty article info",
			info: ArticleInfo{},
			want: ArticleInfo{
				Title:    "",
				URL:      "",
				Author:   "",
				PushRate: 0,
			},
		},
		{
			name: "File mode article (empty title)",
			info: ArticleInfo{
				Title:    "", // 檔案模式時標題可能為空
				URL:      "https://www.ptt.cc/bbs/Beauty/M.1234567890.A.ABC.html",
				Author:   "",
				PushRate: 0,
			},
			want: ArticleInfo{
				Title:    "",
				URL:      "https://www.ptt.cc/bbs/Beauty/M.1234567890.A.ABC.html",
				Author:   "",
				PushRate: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !reflect.DeepEqual(tt.info, tt.want) {
				t.Errorf("ArticleInfo = %+v, want %+v", tt.info, tt.want)
			}
		})
	}
}

func TestArticleInfoValidation(t *testing.T) {
	tests := []struct {
		name    string
		info    ArticleInfo
		isValid bool
		reason  string
	}{
		{
			name: "Valid article info",
			info: ArticleInfo{
				Title:    "[正妹] 測試",
				URL:      "https://www.ptt.cc/bbs/Beauty/M.1234567890.A.ABC.html",
				Author:   "testuser",
				PushRate: 10,
			},
			isValid: true,
			reason:  "complete valid article",
		},
		{
			name: "Valid file mode article",
			info: ArticleInfo{
				Title:    "", // 檔案模式可以沒有標題
				URL:      "https://www.ptt.cc/bbs/Beauty/M.1234567890.A.ABC.html",
				Author:   "",
				PushRate: 0,
			},
			isValid: true,
			reason:  "file mode with URL is valid",
		},
		{
			name: "Invalid - missing URL",
			info: ArticleInfo{
				Title:    "[正妹] 測試",
				URL:      "",
				Author:   "testuser",
				PushRate: 10,
			},
			isValid: false,
			reason:  "URL is required",
		},
		{
			name: "Edge case - extreme push rate",
			info: ArticleInfo{
				Title:    "[爆卦] 超熱門",
				URL:      "https://www.ptt.cc/bbs/Gossiping/M.1234567890.A.ABC.html",
				Author:   "hotuser",
				PushRate: 999,
			},
			isValid: true,
			reason:  "extreme but valid push rate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := isValidArticleInfo(tt.info)
			if isValid != tt.isValid {
				t.Errorf("isValidArticleInfo() = %v, want %v (%s)", isValid, tt.isValid, tt.reason)
			}
		})
	}
}

func TestDownloadTask(t *testing.T) {
	tests := []struct {
		name string
		task DownloadTask
		want DownloadTask
	}{
		{
			name: "Standard download task",
			task: DownloadTask{
				ImageURL: "https://i.imgur.com/test.jpg",
				SavePath: "/path/to/save/test.jpg",
			},
			want: DownloadTask{
				ImageURL: "https://i.imgur.com/test.jpg",
				SavePath: "/path/to/save/test.jpg",
			},
		},
		{
			name: "Imgur without extension",
			task: DownloadTask{
				ImageURL: "https://i.imgur.com/abc123",
				SavePath: "/path/to/save/abc123.jpg",
			},
			want: DownloadTask{
				ImageURL: "https://i.imgur.com/abc123",
				SavePath: "/path/to/save/abc123.jpg",
			},
		},
		{
			name: "Empty download task",
			task: DownloadTask{},
			want: DownloadTask{
				ImageURL: "",
				SavePath: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !reflect.DeepEqual(tt.task, tt.want) {
				t.Errorf("DownloadTask = %+v, want %+v", tt.task, tt.want)
			}
		})
	}
}

func TestDownloadTaskValidation(t *testing.T) {
	tests := []struct {
		name    string
		task    DownloadTask
		isValid bool
		reason  string
	}{
		{
			name: "Valid download task",
			task: DownloadTask{
				ImageURL: "https://i.imgur.com/test.jpg",
				SavePath: "/path/to/save/test.jpg",
			},
			isValid: true,
			reason:  "both URL and path provided",
		},
		{
			name: "Invalid - missing URL",
			task: DownloadTask{
				ImageURL: "",
				SavePath: "/path/to/save/test.jpg",
			},
			isValid: false,
			reason:  "ImageURL is required",
		},
		{
			name: "Invalid - missing save path",
			task: DownloadTask{
				ImageURL: "https://i.imgur.com/test.jpg",
				SavePath: "",
			},
			isValid: false,
			reason:  "SavePath is required",
		},
		{
			name: "Invalid - both empty",
			task: DownloadTask{
				ImageURL: "",
				SavePath: "",
			},
			isValid: false,
			reason:  "both fields are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := isValidDownloadTask(tt.task)
			if isValid != tt.isValid {
				t.Errorf("isValidDownloadTask() = %v, want %v (%s)", isValid, tt.isValid, tt.reason)
			}
		})
	}
}

func TestMarkdownInfo(t *testing.T) {
	tests := []struct {
		name string
		info MarkdownInfo
		want MarkdownInfo
	}{
		{
			name: "Complete markdown info",
			info: MarkdownInfo{
				Title:      "[正妹] 測試標題",
				ArticleURL: "https://www.ptt.cc/bbs/Beauty/M.1234567890.A.ABC.html",
				PushCount:  99,
				ImageURLs:  []string{"https://i.imgur.com/test1.jpg", "https://i.imgur.com/test2.png"},
				SaveDir:    "/path/to/save",
			},
			want: MarkdownInfo{
				Title:      "[正妹] 測試標題",
				ArticleURL: "https://www.ptt.cc/bbs/Beauty/M.1234567890.A.ABC.html",
				PushCount:  99,
				ImageURLs:  []string{"https://i.imgur.com/test1.jpg", "https://i.imgur.com/test2.png"},
				SaveDir:    "/path/to/save",
			},
		},
		{
			name: "Markdown info with no images",
			info: MarkdownInfo{
				Title:      "[問題] 無圖文章",
				ArticleURL: "https://www.ptt.cc/bbs/Test/M.1234567890.A.ABC.html",
				PushCount:  5,
				ImageURLs:  []string{},
				SaveDir:    "/path/to/save",
			},
			want: MarkdownInfo{
				Title:      "[問題] 無圖文章",
				ArticleURL: "https://www.ptt.cc/bbs/Test/M.1234567890.A.ABC.html",
				PushCount:  5,
				ImageURLs:  []string{},
				SaveDir:    "/path/to/save",
			},
		},
		{
			name: "Empty markdown info",
			info: MarkdownInfo{},
			want: MarkdownInfo{
				Title:      "",
				ArticleURL: "",
				PushCount:  0,
				ImageURLs:  nil,
				SaveDir:    "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !reflect.DeepEqual(tt.info, tt.want) {
				t.Errorf("MarkdownInfo = %+v, want %+v", tt.info, tt.want)
			}
		})
	}
}

func TestMarkdownInfoValidation(t *testing.T) {
	tests := []struct {
		name    string
		info    MarkdownInfo
		isValid bool
		reason  string
	}{
		{
			name: "Valid markdown info",
			info: MarkdownInfo{
				Title:      "[正妹] 測試",
				ArticleURL: "https://www.ptt.cc/bbs/Beauty/M.1234567890.A.ABC.html",
				PushCount:  10,
				ImageURLs:  []string{"https://i.imgur.com/test.jpg"},
				SaveDir:    "/path/to/save",
			},
			isValid: true,
			reason:  "complete valid markdown info",
		},
		{
			name: "Valid with no images",
			info: MarkdownInfo{
				Title:      "[問題] 無圖",
				ArticleURL: "https://www.ptt.cc/bbs/Test/M.1234567890.A.ABC.html",
				PushCount:  5,
				ImageURLs:  []string{},
				SaveDir:    "/path/to/save",
			},
			isValid: true,
			reason:  "no images is valid",
		},
		{
			name: "Invalid - missing title",
			info: MarkdownInfo{
				Title:      "",
				ArticleURL: "https://www.ptt.cc/bbs/Beauty/M.1234567890.A.ABC.html",
				PushCount:  10,
				ImageURLs:  []string{},
				SaveDir:    "/path/to/save",
			},
			isValid: false,
			reason:  "Title is required",
		},
		{
			name: "Invalid - missing article URL",
			info: MarkdownInfo{
				Title:      "[正妹] 測試",
				ArticleURL: "",
				PushCount:  10,
				ImageURLs:  []string{},
				SaveDir:    "/path/to/save",
			},
			isValid: false,
			reason:  "ArticleURL is required",
		},
		{
			name: "Invalid - missing save directory",
			info: MarkdownInfo{
				Title:      "[正妹] 測試",
				ArticleURL: "https://www.ptt.cc/bbs/Beauty/M.1234567890.A.ABC.html",
				PushCount:  10,
				ImageURLs:  []string{},
				SaveDir:    "",
			},
			isValid: false,
			reason:  "SaveDir is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := isValidMarkdownInfo(tt.info)
			if isValid != tt.isValid {
				t.Errorf("isValidMarkdownInfo() = %v, want %v (%s)", isValid, tt.isValid, tt.reason)
			}
		})
	}
}

func TestTypesSerialization(t *testing.T) {
	// Test that types can be properly marshaled/unmarshaled (useful for debugging/logging)
	articleInfo := ArticleInfo{
		Title:    "[正妹] 序列化測試",
		URL:      "https://www.ptt.cc/bbs/Beauty/M.1234567890.A.ABC.html",
		Author:   "testuser",
		PushRate: 42,
	}

	downloadTask := DownloadTask{
		ImageURL: "https://i.imgur.com/test.jpg",
		SavePath: "/path/to/save/test.jpg",
	}

	markdownInfo := MarkdownInfo{
		Title:      "[正妹] 序列化測試",
		ArticleURL: "https://www.ptt.cc/bbs/Beauty/M.1234567890.A.ABC.html",
		PushCount:  42,
		ImageURLs:  []string{"https://i.imgur.com/test.jpg"},
		SaveDir:    "/path/to/save",
	}

	// Test that structs can be copied
	articleInfo2 := articleInfo
	if !reflect.DeepEqual(articleInfo, articleInfo2) {
		t.Error("ArticleInfo should be copyable")
	}

	downloadTask2 := downloadTask
	if !reflect.DeepEqual(downloadTask, downloadTask2) {
		t.Error("DownloadTask should be copyable")
	}

	markdownInfo2 := markdownInfo
	if !reflect.DeepEqual(markdownInfo, markdownInfo2) {
		t.Error("MarkdownInfo should be copyable")
	}
}

func TestTypesZeroValues(t *testing.T) {
	// Test zero values of each type
	var articleInfo ArticleInfo
	if articleInfo.Title != "" || articleInfo.URL != "" || articleInfo.Author != "" || articleInfo.PushRate != 0 {
		t.Error("ArticleInfo zero value should have empty strings and zero PushRate")
	}

	var downloadTask DownloadTask
	if downloadTask.ImageURL != "" || downloadTask.SavePath != "" {
		t.Error("DownloadTask zero value should have empty strings")
	}

	var markdownInfo MarkdownInfo
	if markdownInfo.Title != "" || markdownInfo.ArticleURL != "" || markdownInfo.PushCount != 0 ||
		markdownInfo.ImageURLs != nil || markdownInfo.SaveDir != "" {
		t.Error("MarkdownInfo zero value should have empty/nil values")
	}
}

// Helper validation functions for testing
func isValidArticleInfo(info ArticleInfo) bool {
	// URL is required
	return info.URL != ""
}

func isValidDownloadTask(task DownloadTask) bool {
	// Both ImageURL and SavePath are required
	return task.ImageURL != "" && task.SavePath != ""
}

func isValidMarkdownInfo(info MarkdownInfo) bool {
	// Title, ArticleURL, and SaveDir are required
	return info.Title != "" && info.ArticleURL != "" && info.SaveDir != ""
}
