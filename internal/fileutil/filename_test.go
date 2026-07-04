package fileutil

import (
	"reflect"
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

func TestImageFileNames(t *testing.T) {
	tests := []struct {
		name string
		urls []string
		want []string
	}{
		{
			name: "無碰撞",
			urls: []string{"https://i.imgur.com/a.jpg", "https://i.imgur.com/b.png"},
			want: []string{"a.jpg", "b.png"},
		},
		{
			name: "不同 URL 同檔名加序號後綴",
			urls: []string{
				"https://host1.com/a.jpg",
				"https://host2.com/a.jpg",
				"https://host3.com/a.jpg",
			},
			want: []string{"a.jpg", "a_2.jpg", "a_3.jpg"},
		},
		{
			name: "後綴與既有檔名再碰撞時跳號",
			urls: []string{
				"https://host1.com/a_2.jpg",
				"https://host2.com/a.jpg",
				"https://host3.com/a.jpg",
			},
			want: []string{"a_2.jpg", "a.jpg", "a_3.jpg"},
		},
		{
			name: "空列表",
			urls: nil,
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ImageFileNames(tt.urls); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ImageFileNames(%v) = %v, want %v", tt.urls, got, tt.want)
			}
		})
	}
}
