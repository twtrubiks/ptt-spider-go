package markdown

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/twtrubiks/ptt-spider-go/types"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name    string
		info    types.MarkdownInfo
		wantErr bool
		check   func(t *testing.T, filePath string)
	}{
		{
			name: "Generate basic markdown",
			info: types.MarkdownInfo{
				Title:      "[正妹] 測試標題",
				ArticleURL: "https://www.ptt.cc/bbs/Beauty/M.1234567890.A.ABC.html",
				PushCount:  99,
				ImageURLs: []string{
					"https://i.imgur.com/test1.jpg",
					"https://example.com/test2.png",
				},
				SaveDir: "", // Will be set to temp dir
			},
			wantErr: false,
			check: func(t *testing.T, filePath string) {
				content, err := os.ReadFile(filePath)
				if err != nil {
					t.Fatalf("Failed to read generated file: %v", err)
				}

				contentStr := string(content)

				// Check title
				if !strings.Contains(contentStr, "# [正妹] 測試標題") {
					t.Error("Generated markdown should contain title")
				}

				// Check article URL
				if !strings.Contains(contentStr, "https://www.ptt.cc/bbs/Beauty/M.1234567890.A.ABC.html") {
					t.Error("Generated markdown should contain article URL")
				}

				// Check push count
				if !strings.Contains(contentStr, "**推文數量**: 99") {
					t.Error("Generated markdown should contain push count")
				}

				// Check image references
				if !strings.Contains(contentStr, "![test1.jpg](./test1.jpg)") {
					t.Error("Generated markdown should contain first image reference")
				}
				if !strings.Contains(contentStr, "![test2.png](./test2.png)") {
					t.Error("Generated markdown should contain second image reference")
				}

				// Check structure
				if !strings.Contains(contentStr, "## 圖片列表") {
					t.Error("Generated markdown should contain image list header")
				}
			},
		},
		{
			name: "Generate markdown with imgur links",
			info: types.MarkdownInfo{
				Title:      "[分享] Imgur 測試",
				ArticleURL: "https://www.ptt.cc/bbs/Test/M.1234567890.A.ABC.html",
				PushCount:  50,
				ImageURLs: []string{
					"https://i.imgur.com/abc123",     // No extension
					"https://i.imgur.com/def456.jpg", // With extension
				},
				SaveDir: "", // Will be set to temp dir
			},
			wantErr: false,
			check: func(t *testing.T, filePath string) {
				content, err := os.ReadFile(filePath)
				if err != nil {
					t.Fatalf("Failed to read generated file: %v", err)
				}

				contentStr := string(content)

				// Check imgur link without extension gets .jpg added
				if !strings.Contains(contentStr, "![abc123.jpg](./abc123.jpg)") {
					t.Error("Imgur link without extension should get .jpg added")
				}

				// Check imgur link with extension stays the same
				if !strings.Contains(contentStr, "![def456.jpg](./def456.jpg)") {
					t.Error("Imgur link with extension should stay the same")
				}
			},
		},
		{
			name: "Generate markdown with no images",
			info: types.MarkdownInfo{
				Title:      "[問題] 無圖文章",
				ArticleURL: "https://www.ptt.cc/bbs/Test/M.1234567890.A.ABC.html",
				PushCount:  5,
				ImageURLs:  []string{},
				SaveDir:    "", // Will be set to temp dir
			},
			wantErr: false,
			check: func(t *testing.T, filePath string) {
				content, err := os.ReadFile(filePath)
				if err != nil {
					t.Fatalf("Failed to read generated file: %v", err)
				}

				contentStr := string(content)

				// Should still have basic structure
				if !strings.Contains(contentStr, "# [問題] 無圖文章") {
					t.Error("Should contain title")
				}
				if !strings.Contains(contentStr, "## 圖片列表") {
					t.Error("Should contain image list header even with no images")
				}

				// Should not contain any image references
				if strings.Contains(contentStr, "![") {
					t.Error("Should not contain any image references")
				}
			},
		},
		{
			name: "Generate markdown with special characters in title",
			info: types.MarkdownInfo{
				Title:      "[正妹] 測試/特殊:字元*標題",
				ArticleURL: "https://www.ptt.cc/bbs/Beauty/M.1234567890.A.ABC.html",
				PushCount:  10,
				ImageURLs: []string{
					"https://example.com/test.jpg",
				},
				SaveDir: "", // Will be set to temp dir
			},
			wantErr: false,
			check: func(t *testing.T, filePath string) {
				content, err := os.ReadFile(filePath)
				if err != nil {
					t.Fatalf("Failed to read generated file: %v", err)
				}

				contentStr := string(content)

				// Should preserve special characters in markdown content
				if !strings.Contains(contentStr, "[正妹] 測試/特殊:字元*標題") {
					t.Error("Should preserve special characters in title")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for each test
			tmpDir := t.TempDir()
			tt.info.SaveDir = tmpDir

			err := Generate(tt.info)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Check that README.md was created
				filePath := filepath.Join(tmpDir, "README.md")
				if _, err := os.Stat(filePath); os.IsNotExist(err) {
					t.Error("README.md file was not created")
					return
				}

				// Run custom check if provided
				if tt.check != nil {
					tt.check(t, filePath)
				}
			}
		})
	}
}

func TestGenerateDirectoryCreation(t *testing.T) {
	tmpDir := t.TempDir()

	// Test creating nested directories
	nestedDir := filepath.Join(tmpDir, "level1", "level2", "level3")

	info := types.MarkdownInfo{
		Title:      "[測試] 目錄建立",
		ArticleURL: "https://www.ptt.cc/bbs/Test/M.1234567890.A.ABC.html",
		PushCount:  1,
		ImageURLs:  []string{"https://example.com/test.jpg"},
		SaveDir:    nestedDir,
	}

	err := Generate(info)
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	// Check that nested directory was created
	if _, err := os.Stat(nestedDir); os.IsNotExist(err) {
		t.Error("Nested directory was not created")
	}

	// Check that README.md was created in nested directory
	filePath := filepath.Join(nestedDir, "README.md")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("README.md was not created in nested directory")
	}
}

func TestGenerateFilePermissions(t *testing.T) {
	tmpDir := t.TempDir()

	info := types.MarkdownInfo{
		Title:      "[測試] 檔案權限",
		ArticleURL: "https://www.ptt.cc/bbs/Test/M.1234567890.A.ABC.html",
		PushCount:  1,
		ImageURLs:  []string{},
		SaveDir:    tmpDir,
	}

	err := Generate(info)
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	// Check directory permissions
	dirInfo, err := os.Stat(tmpDir)
	if err != nil {
		t.Fatalf("Failed to stat directory: %v", err)
	}

	// Directory should have 755 permissions (but umask might affect actual permissions)
	// So we check if the directory is at least readable and writable by owner
	dirPerm := dirInfo.Mode().Perm()
	if dirPerm&0700 != 0700 { // Check owner has rwx
		t.Errorf("Directory should be readable and writable by owner, got permissions: %v", dirPerm)
	}

	// Check file permissions
	filePath := filepath.Join(tmpDir, "README.md")
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	// File should have 644 permissions
	expectedFilePerm := os.FileMode(0644)
	if fileInfo.Mode().Perm() != expectedFilePerm {
		t.Errorf("File permissions = %v, want %v", fileInfo.Mode().Perm(), expectedFilePerm)
	}
}

func TestGenerateErrorHandling(t *testing.T) {
	// Test with invalid directory path (only works on Unix-like systems)
	if os.PathSeparator == '/' {
		info := types.MarkdownInfo{
			Title:      "[測試] 錯誤處理",
			ArticleURL: "https://www.ptt.cc/bbs/Test/M.1234567890.A.ABC.html",
			PushCount:  1,
			ImageURLs:  []string{},
			SaveDir:    "/root/invalid/path/that/should/not/be/writable",
		}

		err := Generate(info)
		if err == nil {
			t.Error("Generate() should fail with invalid directory path")
		}

		// Error message should mention directory creation failure
		if !strings.Contains(err.Error(), "建立目錄失敗") {
			t.Errorf("Error message should mention directory creation failure, got: %v", err)
		}
	}
}

func TestGenerateContentFormat(t *testing.T) {
	tmpDir := t.TempDir()

	info := types.MarkdownInfo{
		Title:      "[正妹] 格式測試",
		ArticleURL: "https://www.ptt.cc/bbs/Beauty/M.1234567890.A.ABC.html",
		PushCount:  42,
		ImageURLs: []string{
			"https://i.imgur.com/test1.jpg",
			"https://i.imgur.com/test2",
			"https://example.com/subfolder/test3.png",
		},
		SaveDir: tmpDir,
	}

	err := Generate(info)
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	filePath := filepath.Join(tmpDir, "README.md")
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	lines := strings.Split(string(content), "\n")

	// Check line-by-line structure
	expectedLines := []string{
		"# [正妹] 格式測試",
		"",
		"- **文章網址**: [https://www.ptt.cc/bbs/Beauty/M.1234567890.A.ABC.html](https://www.ptt.cc/bbs/Beauty/M.1234567890.A.ABC.html)",
		"- **推文數量**: 42",
		"",
		"## 圖片列表",
		"",
		"![test1.jpg](./test1.jpg)",
		"![test2.jpg](./test2.jpg)",
		"![test3.png](./test3.png)",
	}

	for i, expectedLine := range expectedLines {
		if i >= len(lines) {
			t.Errorf("Generated content has fewer lines than expected. Missing line %d: %s", i, expectedLine)
			continue
		}
		if lines[i] != expectedLine {
			t.Errorf("Line %d: expected %q, got %q", i, expectedLine, lines[i])
		}
	}
}
