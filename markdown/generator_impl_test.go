package markdown

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/twtrubiks/ptt-spider-go/interfaces"
	"github.com/twtrubiks/ptt-spider-go/types"
)

func TestGeneratorImpl_Generate(t *testing.T) {
	generator := NewGenerator()

	tempDir, err := os.MkdirTemp("", "markdown_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	t.Run("basic markdown generation", func(t *testing.T) {
		testBasicMarkdownGeneration(t, generator, tempDir)
	})

	t.Run("nested directory generation", func(t *testing.T) {
		testNestedDirectoryGeneration(t, generator, tempDir)
	})

	t.Run("no images generation", func(t *testing.T) {
		testNoImagesGeneration(t, generator, tempDir)
	})

	t.Run("error case - invalid directory", func(t *testing.T) {
		testInvalidDirectoryCase(t, generator)
	})
}

func testBasicMarkdownGeneration(t *testing.T, generator interfaces.MarkdownGenerator, tempDir string) {
	info := createTestMarkdownInfo(tempDir)

	err := generator.Generate(info)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	mdPath := filepath.Join(tempDir, "README.md")
	verifyFileExists(t, mdPath)

	content := readGeneratedContent(t, mdPath)
	verifyBasicContent(t, content)
}

func testNestedDirectoryGeneration(t *testing.T, generator interfaces.MarkdownGenerator, tempDir string) {
	nestedDir := filepath.Join(tempDir, "nested", "dir")
	info := createTestMarkdownInfo(nestedDir)

	err := generator.Generate(info)
	if err != nil {
		t.Fatalf("Generate failed for nested directory: %v", err)
	}

	nestedMdPath := filepath.Join(nestedDir, "README.md")
	verifyFileExists(t, nestedMdPath)
}

func testNoImagesGeneration(t *testing.T, generator interfaces.MarkdownGenerator, tempDir string) {
	noImagesDir := filepath.Join(tempDir, "no_images")
	info := createTestMarkdownInfo(noImagesDir)
	info.ImageURLs = []string{}

	err := generator.Generate(info)
	if err != nil {
		t.Fatalf("Generate failed for no images: %v", err)
	}

	noImagesPath := filepath.Join(noImagesDir, "README.md")
	content := readGeneratedContent(t, noImagesPath)
	verifyNoImagesContent(t, content)
}

func testInvalidDirectoryCase(t *testing.T, generator interfaces.MarkdownGenerator) {
	if os.Getuid() == 0 { // Skip if running as root
		t.Skip("Skipping invalid directory test when running as root")
	}

	invalidDir := "/invalid/path/that/should/not/exist"
	info := createTestMarkdownInfo(invalidDir)

	err := generator.Generate(info)
	if err == nil {
		t.Error("Expected error for invalid directory, got nil")
	}
}

func createTestMarkdownInfo(saveDir string) types.MarkdownInfo {
	return types.MarkdownInfo{
		Title:      "測試文章標題",
		ArticleURL: "https://www.ptt.cc/bbs/Beauty/M.1234567890.A.ABC.html",
		PushCount:  100,
		ImageURLs: []string{
			"https://i.imgur.com/test1.jpg",
			"https://i.imgur.com/test2.png",
			"https://imgur.com/abcd123",
		},
		SaveDir: saveDir,
	}
}

func verifyFileExists(t *testing.T, filePath string) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatalf("README.md was not created at %s", filePath)
	}
}

func readGeneratedContent(t *testing.T, filePath string) string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read generated markdown: %v", err)
	}
	return string(content)
}

func verifyBasicContent(t *testing.T, content string) {
	expectations := map[string]string{
		"# 測試文章標題":                                              "Generated markdown should contain title",
		"https://www.ptt.cc/bbs/Beauty/M.1234567890.A.ABC.html": "Generated markdown should contain article URL",
		"**推文數量**: 100":                                         "Generated markdown should contain push count",
		"## 圖片列表":                                               "Generated markdown should contain images section",
		"![test1.jpg](./test1.jpg)":                             "Generated markdown should contain first image",
		"![test2.png](./test2.png)":                             "Generated markdown should contain second image",
		"![abcd123.jpg](./abcd123.jpg)":                         "Generated markdown should contain imgur image with .jpg extension",
	}

	for expected, errorMsg := range expectations {
		if !strings.Contains(content, expected) {
			t.Error(errorMsg)
		}
	}
}

func verifyNoImagesContent(t *testing.T, content string) {
	if !strings.Contains(content, "## 圖片列表") {
		t.Error("Generated markdown should still contain images section header")
	}
}

// TestGeneratorImpl_FileNameCollision 驗證不同 URL 推導出相同檔名時，
// README 的圖片連結會使用與 crawler 下載存檔一致的序號後綴檔名。
func TestGeneratorImpl_FileNameCollision(t *testing.T) {
	generator := NewGenerator()
	tempDir := t.TempDir()

	info := types.MarkdownInfo{
		Title:      "檔名碰撞測試",
		ArticleURL: "https://www.ptt.cc/bbs/Beauty/M.1234567890.A.ABC.html",
		PushCount:  10,
		ImageURLs: []string{
			"https://host1.com/a.jpg",
			"https://host2.com/a.jpg",
		},
		SaveDir: tempDir,
	}

	if err := generator.Generate(info); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	content := readGeneratedContent(t, filepath.Join(tempDir, "README.md"))
	if !strings.Contains(content, "![a.jpg](./a.jpg)") {
		t.Error("Generated markdown should contain first image without suffix")
	}
	if !strings.Contains(content, "![a_2.jpg](./a_2.jpg)") {
		t.Error("Generated markdown should contain second image with _2 suffix")
	}
}

func TestNewGenerator(t *testing.T) {
	generator := NewGenerator()
	if generator == nil {
		t.Error("NewGenerator should return a valid generator instance")
	}

	tempDir, err := os.MkdirTemp("", "test_new_generator")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	info := types.MarkdownInfo{
		Title:      "Test",
		ArticleURL: "https://example.com",
		PushCount:  1,
		ImageURLs:  []string{},
		SaveDir:    tempDir,
	}

	err = generator.Generate(info)
	if err != nil {
		t.Errorf("NewGenerator should return a working generator: %v", err)
	}
}
