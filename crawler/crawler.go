package crawler

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/twtrubiks/ptt-spider-go/config"
	"github.com/twtrubiks/ptt-spider-go/markdown"
	"github.com/twtrubiks/ptt-spider-go/ptt"
	"github.com/twtrubiks/ptt-spider-go/types"
)

var (
	// 用於清理檔名中的非法字元
	invalidChars = regexp.MustCompile(`[\/:*?"<>|]`)
)

// Crawler 結構體包含爬蟲的所有狀態和配置.
// 支援看板模式和檔案模式兩種爬取方式.
type Crawler struct {
	Client   *http.Client   // HTTP 客戶端，用於發送請求
	Board    string         // 看板名稱（看板模式時使用）
	Pages    int            // 要爬取的頁數
	PushRate int            // 推文數門檻
	FileURL  string         // 檔案路徑（檔案模式時使用）
	Config   *config.Config // 配置物件
}

// NewCrawler 建立一個新的 Crawler 實例.
// 參數:
//   - board: 看板名稱
//   - pages: 要爬取的頁數
//   - pushRate: 推文數門檻
//   - fileURL: 包含文章 URL 的檔案路徑（為空時使用看板模式）
//   - cfg: 配置物件
func NewCrawler(board string, pages, pushRate int, fileURL string, cfg *config.Config) (*Crawler, error) {
	client, err := ptt.NewClientWithConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("建立 client 失敗: %w", err)
	}
	return &Crawler{
		Client:   client,
		Board:    board,
		Pages:    pages,
		PushRate: pushRate,
		FileURL:  fileURL,
		Config:   cfg,
	}, nil
}

// Run 啟動爬蟲主程序，採用 Producer-Consumer 架構.
//
// 架構說明:
// 1. articleProducer: 產生文章 URL 列表
// 2. contentParser: 解析文章內容並提取圖片 URL（多個並行）
// 3. downloadWorker: 下載圖片檔案（多個並行）
// 4. markdownWorker: 生成 Markdown 檔案（單個）
//
// 支援優雅關閉，當 context 被取消時會停止所有 worker.
func (c *Crawler) Run(ctx context.Context) {
	startTime := time.Now()
	log.Println("爬蟲啟動...")
	rand.Seed(time.Now().UnixNano()) // 初始化隨機數種子

	// 建立帶有 context 的 channels（使用配置值）
	articleInfoChan := make(chan types.ArticleInfo, c.Config.Crawler.Channels.ArticleInfo)
	downloadTaskChan := make(chan types.DownloadTask, c.Config.Crawler.Channels.DownloadTask)
	markdownTaskChan := make(chan types.MarkdownInfo, c.Config.Crawler.Channels.MarkdownTask)

	var parsersWg, downloadersWg, markdownWg sync.WaitGroup

	// 啟動下載工人池（使用配置值）
	numWorkers := c.Config.Crawler.Workers
	downloadersWg.Add(numWorkers)
	for i := 1; i <= numWorkers; i++ {
		go c.downloadWorker(ctx, i, downloadTaskChan, &downloadersWg)
	}

	// 啟動 Markdown 文件產生工人
	markdownWg.Add(1)
	go c.markdownWorker(ctx, markdownTaskChan, &markdownWg)

	// 啟動內容解析器（使用配置值）
	parserCount := c.Config.Crawler.ParserCount
	parsersWg.Add(parserCount)
	for i := 0; i < parserCount; i++ {
		go c.contentParser(ctx, &parsersWg, articleInfoChan, downloadTaskChan, markdownTaskChan)
	}

	// 啟動一個 goroutine，等待所有文章解析完成後，關閉下載和 Markdown 任務 channel
	go func() {
		parsersWg.Wait()
		close(downloadTaskChan)
		close(markdownTaskChan)
	}()

	// 根據模式產生文章任務
	// 這部分是阻塞的，會等到所有文章 URL 都被找到並發送到 articleInfoChan
	if c.FileURL != "" {
		c.articleProducerFromFile(ctx, articleInfoChan)
	} else {
		c.articleProducer(ctx, articleInfoChan)
	}

	// 等待所有下載和 Markdown 任務完成
	downloadersWg.Wait()
	markdownWg.Wait()

	// 檢查是否因為 context 取消而結束
	if ctx.Err() != nil {
		log.Printf("爬蟲因中斷信號而結束，總耗時: %s", time.Since(startTime))
	} else {
		log.Printf("爬蟲結束，總耗時: %s", time.Since(startTime))
	}
}

// articleProducer 產生文章資訊到 channel
func (c *Crawler) articleProducer(ctx context.Context, articleInfoChan chan<- types.ArticleInfo) {
	defer close(articleInfoChan)

	maxPage, err := ptt.GetMaxPage(ctx, c.Client, c.Board)
	if err != nil {
		if ctx.Err() != nil {
			log.Printf("獲取最大頁數時被中斷: %v", ctx.Err())
			return
		}
		log.Fatalf("獲取最大頁數失敗: %v", err)
	}

	log.Printf("看板 %s 最大頁數為: %d", c.Board, maxPage)

	for i := 0; i < c.Pages; i++ {
		// 檢查 context 是否已取消
		select {
		case <-ctx.Done():
			log.Println("文章列表爬取被中斷")
			return
		default:
		}

		currentPage := maxPage - i
		pageURL := fmt.Sprintf("%s/bbs/%s/index%d.html", ptt.PttHead, c.Board, currentPage)
		log.Printf("正在爬取看板列表: %s", pageURL)

		req, err := http.NewRequestWithContext(ctx, "GET", pageURL, nil)
		if err != nil {
			log.Printf("建立請求失敗: %s, 錯誤: %v", pageURL, err)
			continue
		}

		resp, err := c.Client.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				log.Println("列表頁爬取被中斷")
				return
			}
			log.Printf("爬取列表頁失敗: %s, 錯誤: %v", pageURL, err)
			continue
		}
		defer resp.Body.Close()

		articles, err := ptt.ParseArticles(resp.Body)
		if err != nil {
			log.Printf("解析列表頁失敗: %s, 錯誤: %v", pageURL, err)
			continue
		}

		for _, article := range articles {
			if article.PushRate >= c.PushRate {
				select {
				case <-ctx.Done():
					log.Println("文章列表發送被中斷")
					return
				case articleInfoChan <- article:
				}
			}
		}
	}
}

// contentParser 從文章資訊解析內容，並分派下載和 Markdown 任務
func (c *Crawler) contentParser(ctx context.Context, wg *sync.WaitGroup, articleInfoChan <-chan types.ArticleInfo, downloadTaskChan chan<- types.DownloadTask, markdownTaskChan chan<- types.MarkdownInfo) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			log.Println("內容解析器收到中斷信號")
			return
		case article, ok := <-articleInfoChan:
			if !ok {
				return
			}

			// 如果 article.Title 為空，表示是從檔案模式來的，需要解析標題
			logMsg := article.Title
			if logMsg == "" {
				logMsg = article.URL
			}
			log.Printf("正在解析文章: %s", logMsg)

			// 在延遲前檢查 context（使用配置值）
			minDelay, maxDelay := c.Config.GetDelayRange()
			delayRange := int(maxDelay - minDelay)
			delay := minDelay + time.Duration(rand.Intn(delayRange/int(time.Millisecond)))*time.Millisecond
			select {
			case <-ctx.Done():
				log.Println("內容解析器在延遲時被中斷")
				return
			case <-time.After(delay):
			}

			req, err := http.NewRequestWithContext(ctx, "GET", article.URL, nil)
			if err != nil {
				log.Printf("建立文章請求失敗: %s, 錯誤: %v", article.URL, err)
				continue
			}

			resp, err := c.Client.Do(req)
			if err != nil {
				if ctx.Err() != nil {
					log.Println("文章爬取被中斷")
					return
				}
				log.Printf("爬取文章頁失敗: %s, 錯誤: %v", article.URL, err)
				continue
			}

			// 使用新的解析器
			parsedTitle, imgURLs, err := ptt.ParseArticleContent(resp.Body)
			resp.Body.Close() // 確保 Body 被關閉
			if err != nil {
				log.Printf("解析文章頁失敗: %s, 錯誤: %v", article.URL, err)
				continue
			}

			// 決定最終使用的標題
			finalTitle := article.Title
			if c.FileURL != "" && parsedTitle != "" {
				// 在檔案模式下，如果成功解析到標題，則使用它
				finalTitle = parsedTitle
			} else if finalTitle == "" && parsedTitle != "" {
				// 處理從檔案來但未設定標題的情況
				finalTitle = parsedTitle
			}

			if len(imgURLs) > 0 {
				// 使用 finalTitle 來建立資料夾名稱
				dirName := fmt.Sprintf("%s_%d", cleanFileName(finalTitle), article.PushRate)
				saveDir := filepath.Join(c.Board, dirName)

				// 分派下載任務
				for _, imgURL := range imgURLs {
					fileName := filepath.Base(imgURL)
					if strings.Contains(imgURL, "imgur.com") && !strings.Contains(fileName, ".") {
						parsedURL, _ := url.Parse(imgURL)
						fileName = filepath.Base(parsedURL.Path) + ".jpg"
					}

					select {
					case <-ctx.Done():
						log.Println("分派下載任務時被中斷")
						return
					case downloadTaskChan <- types.DownloadTask{
						ImageURL: imgURL,
						SavePath: filepath.Join(saveDir, fileName),
					}:
					}
				}

				// 分派 Markdown 產生任務
				select {
				case <-ctx.Done():
					log.Println("分派 Markdown 任務時被中斷")
					return
				case markdownTaskChan <- types.MarkdownInfo{
					Title:      finalTitle,
					ArticleURL: article.URL,
					PushCount:  article.PushRate,
					ImageURLs:  imgURLs,
					SaveDir:    saveDir,
				}:
				}
			}
		}
	}
}

// markdownWorker 處理 Markdown 檔案的產生
func (c *Crawler) markdownWorker(ctx context.Context, tasks <-chan types.MarkdownInfo, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("Markdown 工人啟動")

	for {
		select {
		case <-ctx.Done():
			log.Println("Markdown 工人收到中斷信號")
			return
		case task, ok := <-tasks:
			if !ok {
				log.Println("Markdown 工人結束")
				return
			}
			log.Printf("正在為文章「%s」產生 Markdown 檔案", task.Title)
			if err := markdown.Generate(task); err != nil {
				log.Printf("產生 Markdown 失敗: %v", err)
			}
		}
	}
}

// cleanFileName 清理檔名中的非法字元
func cleanFileName(name string) string {
	return invalidChars.ReplaceAllString(name, "")
}

// downloadWorker 是 Worker Pool 中的一個工人 Goroutine
func (c *Crawler) downloadWorker(ctx context.Context, id int, tasks <-chan types.DownloadTask, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Printf("下載工人 #%d 啟動", id)

	for {
		select {
		case <-ctx.Done():
			log.Printf("下載工人 #%d 收到中斷信號", id)
			return
		case task, ok := <-tasks:
			if !ok {
				log.Printf("下載工人 #%d 結束", id)
				return
			}

			// *** 解決 429 Too Many Requests 的關鍵 ***
			// 在每次下載前隨機延遲（使用配置值）
			minDelay, maxDelay := c.Config.GetDelayRange()
			delayRange := int(maxDelay - minDelay)
			delay := minDelay + time.Duration(rand.Intn(delayRange/int(time.Millisecond)))*time.Millisecond
			log.Printf("工人 #%d 延遲 %v 後下載: %s", id, delay, task.ImageURL)

			// 使用 select 來處理延遲，以便能響應 context 取消
			select {
			case <-ctx.Done():
				log.Printf("下載工人 #%d 在延遲時被中斷", id)
				return
			case <-time.After(delay):
			}

			req, err := http.NewRequestWithContext(ctx, "GET", task.ImageURL, nil)
			if err != nil {
				log.Printf("工人 #%d 建立請求失敗: %s, 錯誤: %v", id, task.ImageURL, err)
				continue
			}

			resp, err := c.Client.Do(req)
			if err != nil {
				if ctx.Err() != nil {
					log.Printf("下載工人 #%d 下載被中斷", id)
					return
				}
				log.Printf("工人 #%d 下載失敗 (GET): %s, 錯誤: %v", id, task.ImageURL, err)
				continue
			}

			if resp.StatusCode == http.StatusTooManyRequests {
				log.Printf("工人 #%d 遇到 429 Too Many Requests，跳過此次下載: %s", id, task.ImageURL)
				resp.Body.Close() // 即使出錯也要關閉 Body
				continue
			}

			if resp.StatusCode != http.StatusOK {
				log.Printf("工人 #%d 下載失敗 (狀態碼 %d): %s", id, resp.StatusCode, task.ImageURL)
				resp.Body.Close() // 即使出錯也要關閉 Body
				continue
			}

			// 使用立即執行的函式與 defer 來確保資源被釋放
			func() {
				defer resp.Body.Close()

				dir := filepath.Dir(task.SavePath)
				if err := os.MkdirAll(dir, 0755); err != nil {
					log.Printf("工人 #%d 建立目錄失敗: %s, 錯誤: %v", id, dir, err)
					return
				}

				file, err := os.Create(task.SavePath)
				if err != nil {
					log.Printf("工人 #%d 建立檔案失敗: %s, 錯誤: %v", id, task.SavePath, err)
					return
				}
				defer file.Close()

				_, err = io.Copy(file, resp.Body)
				if err != nil {
					log.Printf("工人 #%d 寫入檔案失敗: %s, 錯誤: %v", id, task.SavePath, err)
					return
				}
				log.Printf("工人 #%d 下載完成: %s", id, task.SavePath)
			}()
		}
	}
}

// articleProducerFromFile 從檔案讀取 URL 並產生文章資訊
func (c *Crawler) articleProducerFromFile(ctx context.Context, articleInfoChan chan<- types.ArticleInfo) {
	defer close(articleInfoChan)
	log.Println("啟動檔案模式...")

	file, err := os.Open(c.FileURL)
	if err != nil {
		log.Fatalf("開啟檔案失敗: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// 檢查 context 是否已取消
		select {
		case <-ctx.Done():
			log.Println("檔案讀取被中斷")
			return
		default:
		}

		line := strings.TrimSpace(scanner.Text())
		if strings.Contains(line, "https://www.ptt.cc/bbs/") {
			// 檔案模式下，推文數為 0，因為我們需要下載所有指定的文章
			select {
			case <-ctx.Done():
				log.Println("檔案模式文章發送被中斷")
				return
			case articleInfoChan <- types.ArticleInfo{
				URL:      line,
				PushRate: 0, // 預設值，因為我們不知道推文數
			}:
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("讀取檔案時發生錯誤: %v", err)
	}
}
