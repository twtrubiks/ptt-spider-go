# Go PTT Spider - 高效能 PTT 網路爬蟲

[![Go Version](https://img.shields.io/badge/Go-1.24+-blue.svg)](https://golang.org/)
[![Go Report Card](https://img.shields.io/badge/go_report-A+-brightgreen.svg)](https://goreportcard.com/report/github.com/twtrubiks/ptt-spider-go)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## 📋 專案概述

本專案給想學 Go 的朋友, 我自己也不會 Go, 這個專案都是 AI 寫的, 邊做邊學.

Go PTT Spider 是一個使用 Go 語言編寫的高效能 PTT 網路爬蟲，專門設計用於自動化下載 PTT 文章中的圖片並生成組織化的 Markdown 檔案。該專案採用現代化的並行架構，能夠同時處理多篇文章和圖片下載，大幅提升爬取效率。

本專案為 [PTT_Beauty_Spider](https://github.com/twtrubiks/PTT_Beauty_Spider) Python 版本的 Go 語言重新實現，在保持功能完整性的同時，透過 Go 的並行特性大幅提升了執行效率。

### 🎯 核心特色

- **🚀 高效並行架構**: 採用 Goroutine 和 Channel 實現多工處理，最大化爬取速度
- **🔧 雙模式支援**: 支援看板模式和檔案模式，適應不同使用場景
- **🖼️ 智能圖片處理**: 自動識別和下載多種圖片格式，包含 Imgur 連結處理
- **📝 自動化文件生成**: 為每篇文章生成包含圖片預覽的 Markdown 檔案
- **🛡️ 反爬蟲機制**: 內建延遲機制和模擬瀏覽器行為，避免被封鎖
- **⚡ Context 優雅關閉**: 支援 Ctrl+C 中斷信號，能優雅地停止所有 Goroutine 並清理資源

### 執行畫面

下載過程

![img](https://cdn.imgpile.com/f/TrqOhds_xl.png)

結果

![img](https://cdn.imgpile.com/f/7WrVqan_xl.png)

文字檔

![img](https://cdn.imgpile.com/f/GK8cTDN_xl.png)

## 🚀 使用方法

### 安裝需求

- Go 1.24 或更高版本
- 穩定的網路連線

### 快速開始

- clone 專案

```bash
git clone https://github.com/twtrubiks/ptt-spider-go.git
cd ptt-spider-go
```

- 安裝依賴

```bash
go mod tidy
```

- 執行爬蟲

```bash
go run main.go [參數]
```

### 命令列參數

| 參數 | 類型 | 預設值 | 說明 |
|------|------|--------|------|
| `-board` | string | "beauty" | 看板名稱（支援任意公開看板） |
| `-pages` | int | 3 | 要爬取的頁數（從最新頁開始） |
| `-push` | int | 10 | 推文數門檻（篩選熱門文章） |
| `-file` | string | "" | 文章 URL 檔案路徑（啟用檔案模式） |
| `-config` | string | "config.yaml" | 配置檔案路徑（支援自動降級） |

### 使用範例

#### 爬取 Beauty 看板

```bash
# 爬取 Beauty 看板最新 5 頁，推文數 >= 20
go run main.go -board=beauty -pages=5 -push=20
```

#### 爬取 Gossiping 看板

```bash
# 爬取 Gossiping 看板最新 10 頁，推文數 >= 99（爆文）
go run main.go -board=Gossiping -pages=10 -push=99
```

#### 從檔案爬取

```bash
# 建立 urls.txt 檔案
echo "https://www.ptt.cc/bbs/Beauty/M.1234567890.A.ABC.html" > urls.txt
echo "https://www.ptt.cc/bbs/Beauty/M.0987654321.A.DEF.html" >> urls.txt

# 執行爬蟲
go run main.go -file=urls.txt
```

#### 使用自定義配置

```bash
# 使用自定義配置檔案
go run main.go -config=my-config.yaml -board=beauty -pages=5

# 如果配置檔案不存在，會自動使用預設配置
go run main.go -config=non-existent.yaml -board=beauty -pages=5
```

#### 優雅停止爬蟲

```bash
# 執行爬蟲
go run main.go -board=beauty -pages=5

# 在另一個終端或按 Ctrl+C 來優雅停止
# 程式會顯示: "收到中斷信號，正在優雅關閉爬蟲..."
# 所有 Worker 會完成當前任務後停止
```

## 📁 輸出結果

爬取完成後，會在看板名稱的資料夾下生成以下結構：

```cmd
beauty/
├── 文章標題1_推文數/
│   ├── README.md
│   ├── image1.jpg
│   └── image2.png
└── 文章標題2_推文數/
    ├── README.md
    └── image3.gif
```

每個 `README.md` 檔案包含：

- 文章標題和原始連結
- 推文數統計
- 所有圖片的預覽

## ⚙️ 配置管理

### 配置檔案結構

專案支援 YAML 配置檔案，讓您可以在不重新編譯的情況下調整爬蟲參數：

```yaml
crawler:
  workers: 10          # 並行下載工作者數量
  parserCount: 10      # 內容解析器數量

  channels:            # 通道緩衝區設定
    articleInfo: 100   # 文章資訊通道
    downloadTask: 200  # 下載任務通道
    markdownTask: 100  # Markdown 任務通道

  delays:              # 延遲設定（避免被封鎖）
    minMs: 500         # 最小延遲毫秒數
    maxMs: 2000        # 最大延遲毫秒數

  http:                # HTTP 連線池設定
    timeout: "30s"     # 請求超時時間
    maxIdleConns: 100  # 最大空閒連線數
    maxIdleConnsPerHost: 20      # 每個主機最大空閒連線數
    idleConnTimeout: "90s"       # 空閒連線超時時間
    tlsHandshakeTimeout: "10s"   # TLS 握手超時時間
    expectContinueTimeout: "1s"  # Expect Continue 超時時間
```

### 配置場景範例

#### 保守設定（低速但穩定）

```yaml
crawler:
  workers: 3
  delays:
    minMs: 2000
    maxMs: 5000
```

#### 激進設定（高速爬取）

```yaml
crawler:
  workers: 20
  delays:
    minMs: 100
    maxMs: 500
```

#### 記憶體受限環境

```yaml
crawler:
  workers: 5
  channels:
    articleInfo: 50
    downloadTask: 100
    markdownTask: 50
```

### 配置載入機制

- **自動降級**: 如果配置檔案不存在或解析失敗，自動使用預設配置
- **部分配置**: 可以只配置部分參數，其他參數使用預設值
- **向下相容**: 沒有配置檔案時，程式依然正常運行

## 🏗️ 技術架構

### 模組化設計

```cmd
go-ptt/
├── main.go                 # 程式入口點
├── crawler/
│   └── crawler.go          # 爬蟲核心邏輯
├── ptt/
│   ├── client.go          # HTTP 客戶端配置
│   └── parser.go          # HTML 解析器
├── markdown/
│   └── generator.go       # Markdown 生成器
├── types/
│   └── types.go           # 資料結構定義
├── config/
│   └── config.go          # 配置管理模組
├── go.mod                 # 依賴管理
└── config.yaml            # 主配置檔案
```

### 並行處理架構

專案採用生產者-消費者模式，透過 Channel 進行 Goroutine 間的通訊：

1. **Article Producer**: 負責產生文章 URL 列表
2. **Content Parser**: 解析文章內容並提取圖片 URL（10 個併發）
3. **Download Worker**: 執行圖片下載任務（10 個併發）
4. **Markdown Worker**: 生成 Markdown 檔案（1 個）

## 🔧 核心功能

### 1. 雙模式操作

#### 看板模式

- 爬取指定看板的最新文章
- 依據推文數進行篩選
- 自動獲取看板最大頁數

#### 檔案模式

- 從文字檔讀取文章 URL 列表
- 自動解析文章標題作為資料夾名稱
- 適合批量處理特定文章

### 2. 智能圖片處理

支援的圖片格式：

- `.jpg`, `.jpeg`, `.png`, `.gif`
- Imgur 連結自動處理
- HTTP 自動轉 HTTPS

### 3. 反爬蟲策略

- 設定瀏覽器 User-Agent
- 自動處理 PTT over18 驗證
- 隨機延遲機制（500ms-2s）
- HTTP 429 錯誤處理

### 4. Context 優雅關閉機制

- **信號監聽**: 自動監聽 `SIGINT` (Ctrl+C) 和 `SIGTERM` 信號
- **級聯取消**: Context 取消會傳播到所有 Goroutine
- **資源清理**: 確保所有 HTTP 連線和檔案資源正確關閉
- **進度保護**: 正在下載的檔案會完成後再停止
- **狀態報告**: 清楚顯示因中斷信號而結束的狀態

### 5. HTML 解析功能

- **智能解析**: 使用 goquery 進行高效 DOM 操作
- **價文識別**: 自動繪新價文計算（處理 "99" 等字串）
- **文章過濾**: 根據推文數篩選熱門文章
- **下一頁處理**: 自動找到「上一頁」連結進行連續爬取

## 🔍 技術細節

### 依賴套件

- **[goquery](https://github.com/PuerkitoBio/goquery)**: HTML 解析和 DOM 操作
- **[yaml.v3](https://gopkg.in/yaml.v3)**: YAML 配置檔案解析
- **Go 標準庫**: context, net/http, sync, os/signal 等

### 核心資料型別

```go
// 文章基本資訊
type ArticleInfo struct {
    Title    string  // 文章標題
    URL      string  // 文章連結
    Author   string  // 作者
    PushRate int     // 推文數
}

// 下載任務
type DownloadTask struct {
    ImageURL string  // 圖片 URL
    SavePath string  // 儲存路徑
}

// Markdown 資訊
type MarkdownInfo struct {
    Title      string    // 文章標題
    ArticleURL string    // 原文連結
    PushCount  int       // 推文數
    ImageURLs  []string  // 圖片 URL 列表
    SaveDir    string    // 儲存目錄
}
```

### 並行控制

```go
// Worker Pool 配置（現在可通過 config.yaml 調整）
numWorkers := cfg.Crawler.Workers        // 下載工作者數量
parserCount := cfg.Crawler.ParserCount   // 解析器數量

// Channel 緩衝區設定（現在可通過 config.yaml 調整）
articleInfoChan := make(chan types.ArticleInfo, cfg.Crawler.Channels.ArticleInfo)
downloadTaskChan := make(chan types.DownloadTask, cfg.Crawler.Channels.DownloadTask)
markdownTaskChan := make(chan types.MarkdownInfo, cfg.Crawler.Channels.MarkdownTask)
```

### Context 控制流程

```go
// 主程式設定信號監聽和 Context 控制
func (c *Crawler) Run(ctx context.Context) {
    // 設定信號監聽
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

    // 建立可取消的 Context
    ctx, cancel := context.WithCancel(ctx)
    defer cancel()

    // 監聽中斷信號
    go func() {
        <-sigChan
        log.Println("收到中斷信號，正在優雅關閉爬蟲...")
        cancel()
    }()

    // 所有 Worker 都會檢查 context 狀態
    for {
        select {
        case <-ctx.Done():
            log.Println("Worker 收到中斷信號")
            return
        case task := <-taskChan:
            // 處理任務
        }
    }
}
```

### 錯誤處理

- 網路連線失敗自動跳過
- HTTP 429 錯誤特殊處理
- 檔案寫入失敗記錄但不中斷程式
- Context 取消時優雅退出，避免資源洩漏

## 🧪 測試與調試

### 單元測試

專案包含完整的單元測試計畫，詳細測試內容請參考 [TESTING.md](TESTING.md)。

```bash
# 執行所有測試
go test ./...

# 執行特定套件測試
go test ./tests/

# 詳細輸出
go test -v ./tests/

# 測試覆蓋率
go test -cover ./...

# 生成覆蓋率報告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### 測試檔案結構

```text
├── tests/
│   ├── fixtures/           # 測試資料檔案
│   │   ├── article_content.html
│   │   ├── article_with_images.html
│   │   ├── board_list.html
│   │   └── config_*.yaml
│   └── integration_test.go # 整合測試
├── config/config_test.go   # 配置載入測試
├── ptt/ptt_test.go         # PTT 解析測試
├── crawler/crawler_test.go # 爬蟲邏輯測試
├── markdown/markdown_test.go # Markdown 生成測試
└── types/types_test.go     # 資料結構測試
```

### 效能測試工具

專案提供了 [`benchmark.sh`](benchmark.sh) 腳本，用於自動化測試 HTTP 連線池優化的效能提升：

```bash
# 執行效能基準測試
./benchmark.sh
```

**功能特點**：

- 自動建立測試配置檔案（原始 vs 優化）
- 執行對比測試並計算效能提升百分比
- 驗證下載檔案數量一致性
- 提供詳細的測試報告和建議

**測試輸出範例**：

```text
🚀 HTTP 連線池優化效能測試
==============================
測試參數:
- 看板: beauty
- 頁數: 2
- 推文數門檻: 10

📊 效能比較結果
==============================
原始配置 (連線池未優化): 45秒
優化配置 (HTTP 連線池優化): 28秒
🚀 效能提升: 17秒 (37%)
```

### 調試技巧

```bash
# 開啟詳細日誌
export DEBUG=true
go run main.go -board=beauty -pages=1

# 測試單一文章
echo "https://www.ptt.cc/bbs/Beauty/M.xxxxx.html" > test.txt
go run main.go -file=test.txt

# 使用保守配置測試
go run main.go -config=config_original.yaml -board=beauty -pages=1

# 檢查程式碼品質
golangci-lint run
go vet ./...
```

## 🔧 程式碼品質

### 代碼檢查工具

專案整合了 `golangci-lint` 進行全面的代碼品質檢查：

```bash
# 安裝 golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2

# 運行代碼檢查
golangci-lint run

# 運行基本檢查
go vet ./...
gofmt -l .
```

### 啟用的 Linters

- **errcheck**: 檢查未處理的錯誤
- **gosimple**: 簡化代碼建議
- **govet**: Go vet 檢查
- **ineffassign**: 檢查無效賦值
- **unused**: 檢查未使用的常數、變數、函數等
- **goimports**: 檢查 import 格式
- **misspell**: 檢查拼寫錯誤
- **gofmt**: 檢查 gofmt 格式
- **revive**: 替代 golint 的快速 linter
- **gocyclo**: 循環複雜度檢查
- **goconst**: 檢查可以轉為常數的重複字串
- **godot**: 檢查註解是否以句號結尾
- **nakedret**: 檢查裸返回語句
- **whitespace**: 檢查多餘的空白

### 代碼規範

專案遵循以下 Go 程式碼規範：

1. **命名規範**
   - 使用駝峰命名法（CamelCase）
   - 導出函數和變數使用大寫開頭
   - 私有函數和變數使用小寫開頭

2. **註釋規範**
   - 所有導出的類型、函數都有完整註釋
   - 註釋以類型/函數名稱開頭
   - 註釋以句號結尾

3. **錯誤處理**
   - 所有錯誤都被適當處理
   - 使用 `fmt.Errorf` 包裝錯誤並提供上下文
   - 避免忽略錯誤返回值

4. **並發安全**
   - 正確使用 `sync.WaitGroup` 管理 Goroutine
   - 使用 Channel 進行 Goroutine 間通訊
   - 避免共享可變狀態

### 文檔完整性

- **package 文檔**: `doc.go` 提供完整的套件說明
- **結構體文檔**: 所有導出的結構體都有詳細說明
- **函數文檔**: 包含參數說明和返回值說明
- **使用範例**: README.md 包含豐富的使用範例

### 📚 學習資源

#### 相關專案學習路徑

1. **[PTT_Beauty_Spider](https://github.com/twtrubiks/PTT_Beauty_Spider)**: 原始 Python 版本
   - 本專案的 Python 實現版本
   - 功能對比和架構參考
   - 詳細的性能比較請參考 [PTT_Spider_Comparison.md](PTT_Spider_Comparison.md)

2. **[hello-world](https://github.com/twtrubiks/golang-notes/tree/main/hello-world)**: Go 語言基礎學習以及 VSCode 環境設置和工具使用

3. **[go-downloader](https://github.com/twtrubiks/golang-notes/tree/main/go-downloader)**: 簡單的下載實現

4. **本專案進階學習重點**:
   - Producer-Consumer 架構模式
   - Worker Pool 設計模式
   - Context 生命週期管理
   - HTTP 客戶端優化技巧

#### 專案設計重點

**核心功能**:

1. **網頁爬取**: 下載 PTT 網頁內容
2. **內容解析**: 提取文章資訊和圖片連結
3. **圖片下載**: 並行下載多張圖片
4. **Markdown 生成**: 為每篇文章建立索引

**執行流程**:

1. 產生文章 URL 列表（看板模式/檔案模式）
2. 並行解析文章內容
3. 提取圖片 URL 並加入下載佇列
4. Worker Pool 執行圖片下載
5. 生成 Markdown 文件索引

## 📄 授權條款

本專案採用 MIT 授權條款 - 詳見 [LICENSE](LICENSE) 檔案。

## 🛠️ 故障排除

### 常見問題

1. **HTTP 429 錯誤**
   - 增加延遲設定：調整 `config.yaml` 中的 `minMs` 和 `maxMs`
   - 減少 Worker 數量：降低 `workers` 和 `parserCount`

2. **記憶體使用過高**
   - 減少 Channel 緩衝區大小
   - 降低並行 Worker 數量

3. **下載失敗**
   - 檢查網路連線
   - 確認圖片 URL 是否有效
   - 查看是否被目標網站封鎖

## 🎉 致謝

- 感謝 [PTT](https://www.ptt.cc/) 提供豐富的內容平台
- 感謝 [goquery](https://github.com/PuerkitoBio/goquery) 提供強大的 HTML 解析功能
- 感謝所有為這個專案做出貢獻的開發者們！

---

## ⚠️ 免責聲明

- 本工具僅供學習和研究用途
- 請遵守 PTT 的使用條款和相關法律規定
- 使用時請適度控制爬取頻率，避免對伺服器造成過大負擔
- 請勿用於商業用途或大規模資料採集
- 使用者需自行承擔使用本工具的相關責任

---

## Donation

文章都是我自己研究內化後原創，如果有幫助到您，也想鼓勵我的話，歡迎請我喝一杯咖啡  :laughing:

綠界科技ECPAY ( 不需註冊會員 )

![alt tag](https://payment.ecpay.com.tw/Upload/QRCode/201906/QRCode_672351b8-5ab3-42dd-9c7c-c24c3e6a10a0.png)

[贊助者付款](http://bit.ly/2F7Jrha)

歐付寶 ( 需註冊會員 )

![alt tag](https://i.imgur.com/LRct9xa.png)

[贊助者付款](https://payment.opay.tw/Broadcaster/Donate/9E47FDEF85ABE383A0F5FC6A218606F8)

## 贊助名單

[贊助名單](https://github.com/twtrubiks/Thank-you-for-donate)
