# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

PTT Spider Go 是一個高效能的 PTT 網頁爬蟲，使用 Go 的 goroutine 並行架構爬取文章、下載圖片並自動產生 Markdown 文件。支援看板模式與檔案模式兩種運作方式。

## Build & Development Commands

```bash
# 建置
go build ./...

# 執行（看板模式）
go run main.go -board beauty -pages 3 -push 10

# 執行（檔案模式）
go run main.go -file urls.txt -board beauty

# TUI 互動模式
go run main.go -tui

# 帶設定檔執行
go run main.go -config config.yaml

# 測試
go test ./...                                    # 全部測試
go test -race ./...                              # 含 race detection
go test -v ./crawler/                            # 單一套件
go test -v -run TestRetryOn429 ./crawler/        # 單一測試
go test -coverprofile=coverage.out ./...         # 覆蓋率報告
go tool cover -html=coverage.out                 # 瀏覽器查看覆蓋率

# 程式碼品質
golangci-lint run                                # Linter（設定在 .golangci.yml）
go vet ./...
gofmt -s -w .
```

## Architecture

### Producer-Consumer 並行架構

```
articleProducer (1 goroutine)
    → ArticleInfo channel (buffer: 100)
        → contentParser (10 goroutines)
            → DownloadTask channel (buffer: 200)
                → downloadWorker (10 goroutines) → 儲存圖片
            → MarkdownInfo channel (buffer: 100)
                → markdownWorker (1 goroutine) → 產生 README.md

各階段透過 emit() → ProgressEvent channel → TUI 即時進度畫面（可選）
```

所有 worker 透過 `context.Context` 實現優雅關閉，channel 關閉順序由 `waitAndCleanup()` 協調。

### 套件職責

| 套件 | 職責 |
|------|------|
| `crawler` | 核心協調器：Producer-Consumer 流程、worker pool、HTTP 429 重試 (`retry.go`) |
| `ptt` | PTT 網站整合：HTTP client（含連線池和 Over18 cookie）、HTML 解析（goquery） |
| `interfaces` | 3 個核心介面：`HTTPClient`、`Parser`、`MarkdownGenerator` |
| `types` | 資料結構：`ArticleInfo`、`DownloadTask`、`MarkdownInfo`、`ProgressEvent` |
| `config` | YAML 設定載入，失敗時自動降級為預設值 |
| `errors` | 5 種結構化錯誤型別，支援 `errors.As`/`errors.Is` |
| `markdown` | 為每篇文章產生帶圖片連結的 Markdown 檔案 |
| `performance` | 記憶體和 goroutine 監控 |
| `mocks` | Function field pattern 的 mock 物件（無外部 mock 框架） |
| `internal/ioutil` | `CloseWithLog` 統一資源關閉 |
| `ui` | `Logger` 介面與實作：`PlainLogger`（純文字）、`StyledLogger`（Lip Gloss 彩色輸出）、`NoopLogger`（靜默）；TUI 互動式啟動表單（`huh`）；即時進度 TUI（Bubble Tea） |

### 依賴注入

- `NewCrawler()` — 正式使用，自動建立所有依賴，支援 `...Option` variadic 參數
- `NewCrawlerWithDependencies()` — 測試用，注入 mock 物件

### Option Pattern

`NewCrawler` 支援可選配置：

```go
crawler.WithProgress(ch)  // 注入進度事件 channel（TUI 即時進度用）
crawler.WithLogger(l)     // 注入自訂 Logger
```

### Mock 模式

使用 function field pattern（非 testify），mock 定義在 `mocks/mocks.go`：

```go
mockParser := &mocks.MockParser{
    ParseMaxPageFunc: func(body io.Reader) (int, error) {
        return 5, nil
    },
}
```

### 設定檔 (config.yaml)

關鍵設定項：`workers`（下載並行數）、`parserCount`（解析並行數）、`channels`（buffer 大小）、`delays`（反爬蟲延遲 ms）、`http`（連線池參數）。

## 開發原則

- 簡單方案優先，不要 over-engineer
- README 不要用過多 emoji 和行銷語調，保持開發者風格

### 可逆性與回滾優先

- 保持變更易於還原（小範圍、小型 commit、清晰的影響範圍）
- 對於有風險的變更，在合併前定義回滾路徑
- 避免阻礙安全回滾的混合巨型補丁
- 實作前用程式碼搜尋驗證假設
- 優先選擇確定性行為而非聰明的捷徑
- 如果不確定，留下帶有驗證脈絡的具體 TODO，而非隱藏的猜測

### 反模式（禁止事項）

- 不要為了小便利而新增重量級相依
- 不要靜默弱化安全策略或存取約束
- 不要「以防萬一」地新增推測性的設定/功能旗標
- 不要將大量純格式化變更與功能變更混合
- 不要「順便」修改無關模組
- 不要在沒有明確說明的情況下繞過失敗的檢查
- 不要在重構 commit 中隱藏改變行為的副作用
- 不要在測試資料、範例、文件或 commit 中包含個人身分或敏感資訊
- 除非維護者明確要求，否則不要嘗試儲存庫品牌重塑/身分替換
- 除非維護者明確要求，否則不要引入新的平台面（例如 `web` 應用、儀表板、UI 入口）

## Key Conventions

- Go 最低版本：1.26
- 使用 `time.NewTimer()` 而非 `time.After()`（避免 timer 洩漏）
- 使用 `math/rand/v2` 的 `rand.IntN()`（避免全域鎖競爭）
- 使用 `ioutil.CloseWithLog()` 關閉所有 `io.Closer` 資源
- 錯誤型別使用不可變模式（`WithContext()` 回傳新副本）
- `startProducer` 必須在 goroutine 中執行（避免 context 取消時 deadlock）
