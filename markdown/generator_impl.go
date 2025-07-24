package markdown

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/twtrubiks/ptt-spider-go/constants"
	"github.com/twtrubiks/ptt-spider-go/errors"
	"github.com/twtrubiks/ptt-spider-go/interfaces"
	"github.com/twtrubiks/ptt-spider-go/types"
)

// GeneratorImpl 實現 MarkdownGenerator 介面
type GeneratorImpl struct{}

// NewGenerator 建立新的 Markdown 生成器實例
func NewGenerator() interfaces.MarkdownGenerator {
	return &GeneratorImpl{}
}

// Generate 實現 MarkdownGenerator 介面的 Generate 方法
func (g *GeneratorImpl) Generate(info types.MarkdownInfo) error {
	// 確保儲存目錄存在
	if err := os.MkdirAll(info.SaveDir, constants.DirPermission); err != nil {
		return errors.NewFileError(fmt.Sprintf("建立目錄失敗 %s", info.SaveDir), err)
	}

	// 設定 Markdown 檔案路徑
	mdFileName := "README.md"
	mdFilePath := filepath.Join(info.SaveDir, mdFileName)

	// 建立一個 strings.Builder 來高效地建立 Markdown 內容
	var builder strings.Builder

	// 寫入標題
	builder.WriteString(fmt.Sprintf("# %s\n\n", info.Title))

	// 寫入文章資訊
	builder.WriteString(fmt.Sprintf("- **文章網址**: [%s](%s)\n", info.ArticleURL, info.ArticleURL))
	builder.WriteString(fmt.Sprintf("- **推文數量**: %d\n\n", info.PushCount))

	// 寫入圖片標題
	builder.WriteString("## 圖片列表\n\n")

	// 寫入圖片連結
	for _, imgURL := range info.ImageURLs {
		// 使用相對路徑來引用本地圖片
		imgFileName := filepath.Base(imgURL)
		// 處理沒有副檔名的 imgur 連結檔名
		if strings.Contains(imgURL, "imgur.com") && !strings.Contains(imgFileName, ".") {
			imgFileName += ".jpg"
		}

		// Markdown 格式：![替代文字](圖片路徑)
		builder.WriteString(fmt.Sprintf("![%s](./%s)\n", imgFileName, imgFileName))
	}

	// 將組合好的內容寫入檔案
	err := os.WriteFile(mdFilePath, []byte(builder.String()), constants.FilePermission)
	if err != nil {
		return errors.NewFileError("寫入 Markdown 檔案失敗", err)
	}

	return nil
}
