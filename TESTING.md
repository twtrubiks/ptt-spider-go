# 測試計畫 - Go PTT Spider

## 📋 測試覆蓋現狀

目前專案缺乏單元測試，建議為每個套件添加測試檔案，特別是 parser 和 crawler 核心邏輯。

## 🎯 基本測試策略

### 1. 測試優先順序

#### A. **高優先級** (核心功能)
- **Config 套件**: 配置載入、預設值、錯誤處理
- **PTT 套件**: HTML 解析、文章列表、圖片 URL 擷取

#### B. **中優先級** (整合功能)
- **Crawler 套件**: 工作池管理、Context 取消、檔名清理
- **Markdown 套件**: 檔案生成、目錄建立

#### C. **低優先級** (資料結構)
- **Types 套件**: 結構體驗證、資料完整性

### 2. 測試檔案結構

```
go-ptt/ptt-spider-go/
├── tests/
│   ├── fixtures/               # 測試資料
│   │   ├── board_list.html     # 看板列表頁面
│   │   ├── article_content.html # 文章內容頁面
│   │   ├── article_with_images.html # 含圖片的文章
│   │   ├── config_valid.yaml   # 有效配置檔案
│   │   └── config_invalid.yaml # 無效配置檔案
│   ├── config_test.go          # 配置載入測試
│   ├── ptt_test.go             # PTT 解析測試
│   ├── crawler_test.go         # 爬蟲邏輯測試
│   ├── markdown_test.go        # Markdown 生成測試
│   └── integration_test.go     # 整合測試
└── [現有檔案結構]
```

## 🔧 詳細測試計畫

### A. Config 套件測試 (config_test.go)

**測試重點**:
- 配置檔案載入成功
- 配置檔案不存在時使用預設配置
- 配置檔案損壞時的降級機制
- 超時時間解析正確性
- 延遲範圍計算準確性
- HTTP 連線池配置解析

**測試案例**:
```go
func TestConfigLoad(t *testing.T) {
    // 測試正常載入
    // 測試檔案不存在
    // 測試檔案損壞
}

func TestDefaultConfig(t *testing.T) {
    // 測試預設配置值
}

func TestGetTimeoutDuration(t *testing.T) {
    // 測試超時解析
}

func TestGetDelayRange(t *testing.T) {
    // 測試延遲範圍
}

// HTTP 連線池測試
func TestHTTPConfigParsing(t *testing.T) {
    // 測試 HTTP 配置解析
    // 測試 IdleConnTimeout 解析
    // 測試 TLSHandshakeTimeout 解析
    // 測試 ExpectContinueTimeout 解析
}

func TestGetIdleConnTimeout(t *testing.T) {
    // 測試空閒連線超時解析
}

func TestGetTLSHandshakeTimeout(t *testing.T) {
    // 測試 TLS 握手超時解析
}

func TestGetExpectContinueTimeout(t *testing.T) {
    // 測試 Expect Continue 超時解析
}
```

### B. PTT 套件測試 (ptt_test.go)

**測試重點**:
- HTML 解析正確性
- 文章列表擷取
- 圖片 URL 識別和清理
- 推文數解析 (包含 "爆" 和 "X" 前綴)
- 錯誤頁面處理
- HTTP 客戶端建立和連線池配置

**測試案例**:
```go
func TestParseArticles(t *testing.T) {
    // 測試正常文章列表解析
    // 測試包含刪除文章的列表
    // 測試推文數解析 (正常、"爆"、"X1")
}

func TestParseArticleContent(t *testing.T) {
    // 測試標題擷取
    // 測試圖片 URL 擷取
    // 測試各種圖片格式
    // 測試 Imgur 連結處理
}

func TestGetMaxPage(t *testing.T) {
    // 測試最大頁數擷取
}

func TestNewClient(t *testing.T) {
    // 測試基本 HTTP 客戶端建立
}

// HTTP 連線池客戶端測試
func TestNewClientWithConfig(t *testing.T) {
    // 測試使用配置的 HTTP 客戶端建立
    // 測試連線池參數設定
    // 測試超時配置
    // 測試 Transport 配置正確性
}

func TestHTTPTransportConfiguration(t *testing.T) {
    // 測試 HTTP Transport 配置
    // 測試 MaxIdleConns 設定
    // 測試 MaxIdleConnsPerHost 設定
    // 測試各種超時設定
}

func TestConnectionPoolBehavior(t *testing.T) {
    // 測試連線池行為
    // 測試連線復用
    // 測試連線超時
}
```

### C. Crawler 套件測試 (crawler_test.go)

**測試重點**:
- Crawler 建立和初始化
- 工作池管理
- Context 取消機制
- 檔名清理函式
- 錯誤處理

**測試案例**:
```go
func TestNewCrawler(t *testing.T) {
    // 測試 Crawler 建立
}

func TestCleanFileName(t *testing.T) {
    // 測試檔名清理
}

func TestCrawlerWithContext(t *testing.T) {
    // 測試 Context 取消
}
```

### D. Markdown 套件測試 (markdown_test.go)

**測試重點**:
- Markdown 檔案生成
- 目錄建立
- 檔案內容正確性
- 圖片連結格式

**測試案例**:
```go
func TestGenerateMarkdown(t *testing.T) {
    // 測試 Markdown 生成
}

func TestCreateDirectory(t *testing.T) {
    // 測試目錄建立
}
```

## 🧪 測試資料準備

### 1. HTML 測試資料

**board_list.html** - 看板列表頁面範例:
```html
<div class="r-ent">
    <div class="nrec"><span class="hl f2">爆</span></div>
    <div class="title">
        <a href="/bbs/Beauty/M.1234567890.A.ABC.html">[正妹] 測試標題</a>
    </div>
    <div class="author">testuser</div>
    <div class="date">12/25</div>
</div>
```

**article_content.html** - 文章內容頁面範例:
```html
<div id="main-content">
    <div class="article-metaline">
        <span class="article-meta-tag">標題</span>
        <span class="article-meta-value">[正妹] 測試標題</span>
    </div>
    <div class="article-content">
        測試內容
        <a href="https://i.imgur.com/test.jpg">https://i.imgur.com/test.jpg</a>
        <a href="https://example.com/test.png">https://example.com/test.png</a>
    </div>
</div>
```

### 2. 配置測試資料

**config_valid.yaml**:
```yaml
crawler:
  workers: 5
  delays:
    minMs: 1000
    maxMs: 3000
  # HTTP 連線池配置測試
  http:
    timeout: "30s"
    maxIdleConns: 50
    maxIdleConnsPerHost: 10
    idleConnTimeout: "60s"
    tlsHandshakeTimeout: "10s"
    expectContinueTimeout: "1s"
```


**config_invalid.yaml**:
```yaml
invalid_structure:
  - broken yaml content
```

## 🚀 執行測試

### 基本測試命令

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

### 測試執行範例

```bash
# 執行配置測試
go test -v ./config/

# 執行 PTT 解析測試
go test -v ./ptt/

# 執行 Crawler 測試
go test -v ./crawler/

# 執行整合測試
go test -v ./tests/

# 執行基準測試
go test -bench=. ./tests/

# 執行特定測試案例
go test -v ./config/ -run TestLoad
go test -v ./ptt/ -run TestParseArticles
```

### 預期測試輸出範例

成功的測試執行應該看起來像這樣：

```
$ go test ./...
ok      github.com/twtrubiks/ptt-spider-go/config    0.003s
ok      github.com/twtrubiks/ptt-spider-go/crawler   3.961s
ok      github.com/twtrubiks/ptt-spider-go/markdown  0.003s
ok      github.com/twtrubiks/ptt-spider-go/ptt       1.737s
ok      github.com/twtrubiks/ptt-spider-go/tests     2.683s
ok      github.com/twtrubiks/ptt-spider-go/types     0.002s

$ go test -cover ./...
ok      github.com/twtrubiks/ptt-spider-go/config    coverage: 94.6% of statements
ok      github.com/twtrubiks/ptt-spider-go/crawler   coverage: 38.4% of statements
ok      github.com/twtrubiks/ptt-spider-go/markdown  coverage: 94.4% of statements
ok      github.com/twtrubiks/ptt-spider-go/ptt       coverage: 66.0% of statements
ok      github.com/twtrubiks/ptt-spider-go/tests     coverage: [no statements]
ok      github.com/twtrubiks/ptt-spider-go/types     coverage: [no statements]
```
