package fileutil

import (
	"testing"
)

func TestImageFileName(t *testing.T) {
	tests := []struct {
		name   string
		imgURL string
		want   string
	}{
		{
			name:   "一般圖片 URL",
			imgURL: "https://i.imgur.com/test1.jpg",
			want:   "test1.jpg",
		},
		{
			name:   "png 圖片",
			imgURL: "https://example.com/photos/test2.png",
			want:   "test2.png",
		},
		{
			name:   "imgur 無副檔名連結補上 .jpg",
			imgURL: "https://imgur.com/abcd123",
			want:   "abcd123.jpg",
		},
		{
			name:   "i.imgur 無副檔名連結補上 .jpg",
			imgURL: "https://i.imgur.com/abc123",
			want:   "abc123.jpg",
		},
		{
			name:   "帶 query string 的 URL 只取 path 部分",
			imgURL: "https://example.com/a.jpg?width=100",
			want:   "a.jpg",
		},
		{
			name:   "imgur 無副檔名且帶 query string",
			imgURL: "https://imgur.com/abc?x=1",
			want:   "abc.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ImageFileName(tt.imgURL); got != tt.want {
				t.Errorf("ImageFileName(%q) = %q, want %q", tt.imgURL, got, tt.want)
			}
		})
	}
}
