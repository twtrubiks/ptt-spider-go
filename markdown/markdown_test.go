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
			name:    "Generate basic markdown",
			info:    createBasicMarkdownInfo(),
			wantErr: false,
			check:   checkBasicMarkdown,
		},
		{
			name:    "Generate markdown with imgur links",
			info:    createImgurMarkdownInfo(),
			wantErr: false,
			check:   checkImgurMarkdown,
		},
		{
			name:    "Generate markdown with no images",
			info:    createNoImagesMarkdownInfo(),
			wantErr: false,
			check:   checkNoImagesMarkdown,
		},
		{
			name:    "Generate markdown with special characters in title",
			info:    createSpecialCharsMarkdownInfo(),
			wantErr: false,
			check:   checkSpecialCharsMarkdown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestCase(t, tt)
		})
	}
}

// Test data creation functions
func createBasicMarkdownInfo() types.MarkdownInfo {
	return types.MarkdownInfo{
		Title:      "[正妹] 測試標題",
		ArticleURL: "https://www.ptt.cc/bbs/Beauty/M.1234567890.A.ABC.html",
		PushCount:  99,
		ImageURLs: []string{
			"https://i.imgur.com/test1.jpg",
			"https://example.com/test2.png",
		},
		SaveDir: "",
	}
}

func createImgurMarkdownInfo() types.MarkdownInfo {
	return types.MarkdownInfo{
		Title:      "[分享] Imgur 測試",
		ArticleURL: "https://www.ptt.cc/bbs/Test/M.1234567890.A.ABC.html",
		PushCount:  50,
		ImageURLs: []string{
			"https://i.imgur.com/abc123",
			"https://i.imgur.com/def456.jpg",
		},
		SaveDir: "",
	}
}

func createNoImagesMarkdownInfo() types.MarkdownInfo {
	return types.MarkdownInfo{
		Title:      "[問題] 無圖文章",
		ArticleURL: "https://www.ptt.cc/bbs/Test/M.1234567890.A.ABC.html",
		PushCount:  5,
		ImageURLs:  []string{},
		SaveDir:    "",
	}
}

func createSpecialCharsMarkdownInfo() types.MarkdownInfo {
	return types.MarkdownInfo{
		Title:      "[正妹] 測試/特殊:字元*標題",
		ArticleURL: "https://www.ptt.cc/bbs/Beauty/M.1234567890.A.ABC.html",
		PushCount:  10,
		ImageURLs: []string{
			"https://example.com/test.jpg",
		},
		SaveDir: "",
	}
}

// Check functions
func checkBasicMarkdown(t *testing.T, filePath string) {
	content := readFileContent(t, filePath)

	checkContentContains(t, content, "# [正妹] 測試標題", "Generated markdown should contain title")
	checkContentContains(t, content, "https://www.ptt.cc/bbs/Beauty/M.1234567890.A.ABC.html", "Generated markdown should contain article URL")
	checkContentContains(t, content, "**推文數量**: 99", "Generated markdown should contain push count")
	checkContentContains(t, content, "![test1.jpg](./test1.jpg)", "Generated markdown should contain first image reference")
	checkContentContains(t, content, "![test2.png](./test2.png)", "Generated markdown should contain second image reference")
	checkContentContains(t, content, "## 圖片列表", "Generated markdown should contain image list header")
}

func checkImgurMarkdown(t *testing.T, filePath string) {
	content := readFileContent(t, filePath)

	checkContentContains(t, content, "![abc123.jpg](./abc123.jpg)", "Imgur link without extension should get .jpg added")
	checkContentContains(t, content, "![def456.jpg](./def456.jpg)", "Imgur link with extension should stay the same")
}

func checkNoImagesMarkdown(t *testing.T, filePath string) {
	content := readFileContent(t, filePath)

	checkContentContains(t, content, "# [問題] 無圖文章", "Should contain title")
	checkContentContains(t, content, "## 圖片列表", "Should contain image list header even with no images")

	if strings.Contains(content, "![") {
		t.Error("Should not contain any image references")
	}
}

func checkSpecialCharsMarkdown(t *testing.T, filePath string) {
	content := readFileContent(t, filePath)

	checkContentContains(t, content, "[正妹] 測試/特殊:字元*標題", "Should preserve special characters in title")
}

// Helper functions
func readFileContent(t *testing.T, filePath string) string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}
	return string(content)
}

func checkContentContains(t *testing.T, content, expected, errorMsg string) {
	if !strings.Contains(content, expected) {
		t.Error(errorMsg)
	}
}

func runTestCase(t *testing.T, tt struct {
	name    string
	info    types.MarkdownInfo
	wantErr bool
	check   func(t *testing.T, filePath string)
}) {
	tmpDir := t.TempDir()
	tt.info.SaveDir = tmpDir

	err := Generate(tt.info)
	if (err != nil) != tt.wantErr {
		t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
		return
	}

	if !tt.wantErr {
		filePath := filepath.Join(tmpDir, "README.md")
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Error("README.md file was not created")
			return
		}

		if tt.check != nil {
			tt.check(t, filePath)
		}
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
