package interfaces

import (
	"testing"
)

// TestInterfacesExist 測試所有介面定義是否正確
func TestInterfacesExist(t *testing.T) {
	// 這個測試主要是為了確保介面定義正確，能夠編譯通過
	// 實際的介面測試會在實現這些介面的具體類型中進行

	// 檢查介面是否定義
	var (
		_ HTTPClient        = nil
		_ Parser            = nil
		_ MarkdownGenerator = nil
	)

	// 如果編譯通過，則測試通過
	t.Log("所有介面定義正確")
}
