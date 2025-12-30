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
	"github.com/twtrubiks/ptt-spider-go/constants"
	"github.com/twtrubiks/ptt-spider-go/interfaces"
	"github.com/twtrubiks/ptt-spider-go/markdown"
	"github.com/twtrubiks/ptt-spider-go/performance"
	"github.com/twtrubiks/ptt-spider-go/ptt"
	"github.com/twtrubiks/ptt-spider-go/types"
)

var (
	// 用於清理檔名中的非法字元
	invalidChars = regexp.MustCompile(`[\/:*?"<>|]`)
)

// WorkerChannels 包含所有工人使用的 channels
type WorkerChannels struct {
	ArticleInfo  chan types.ArticleInfo
	DownloadTask chan types.DownloadTask
	MarkdownTask chan types.MarkdownInfo
}

// Workers 包含所有工人的 WaitGroup
type Workers struct {
	Parsers     *sync.WaitGroup
	Downloaders *sync.WaitGroup
	Markdown    *sync.WaitGroup
}

// Crawler 結構體包含爬蟲的所有狀態和配置.
// 支援看板模式和檔案模式兩種爬取方式.
type Crawler struct {
	Client            interfaces.HTTPClient        // HTTP 客戶端，用於發送請求
	Parser            interfaces.Parser            // HTML 解析器
	MarkdownGenerator interfaces.MarkdownGenerator // Markdown 生成器
	Optimizer         *performance.Optimizer       // 效能優化器
	Board             string                       // 看板名稱（看板模式時使用）
	Pages             int                          // 要爬取的頁數
	PushRate          int                          // 推文數門檻
	FileURL           string                       // 檔案路徑（檔案模式時使用）
	Config            *config.Config               // 配置物件
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

	// 建立效能優化器 (記憶體閾值 100MB，檢查間隔 30 秒)
	optimizer := performance.NewOptimizer(100, 30*time.Second)

	return &Crawler{
		Client:            client,
		Parser:            ptt.NewParser(),
		MarkdownGenerator: markdown.NewGenerator(),
		Optimizer:         optimizer,
		Board:             board,
		Pages:             pages,
		PushRate:          pushRate,
		FileURL:           fileURL,
		Config:            cfg,
	}, nil
}

// NewCrawlerWithDependencies 建立一個新的 Crawler 實例，支援依賴注入
func NewCrawlerWithDependencies(
	client interfaces.HTTPClient,
	parser interfaces.Parser,
	markdownGen interfaces.MarkdownGenerator,
	board string,
	pages, pushRate int,
	fileURL string,
	cfg *config.Config,
) *Crawler {
	// 建立效能優化器
	optimizer := performance.NewOptimizer(100, 30*time.Second)

	return &Crawler{
		Client:            client,
		Parser:            parser,
		MarkdownGenerator: markdownGen,
		Optimizer:         optimizer,
		Board:             board,
		Pages:             pages,
		PushRate:          pushRate,
		FileURL:           fileURL,
		Config:            cfg,
	}
}

// initializeChannels 初始化所有工人使用的 channels
func (c *Crawler) initializeChannels() *WorkerChannels {
	return &WorkerChannels{
		ArticleInfo:  make(chan types.ArticleInfo, c.Config.Crawler.Channels.ArticleInfo),
		DownloadTask: make(chan types.DownloadTask, c.Config.Crawler.Channels.DownloadTask),
		MarkdownTask: make(chan types.MarkdownInfo, c.Config.Crawler.Channels.MarkdownTask),
	}
}

// startWorkers 啟動所有工人並返回 WaitGroup
func (c *Crawler) startWorkers(ctx context.Context, channels *WorkerChannels) *Workers {
	var parsersWg, downloadersWg, markdownWg sync.WaitGroup

	// 啟動下載工人池
	numWorkers := c.Config.Crawler.Workers
	downloadersWg.Add(numWorkers)
	for i := 1; i <= numWorkers; i++ {
		go c.downloadWorker(ctx, i, channels.DownloadTask, &downloadersWg)
	}

	// 啟動 Markdown 文件產生工人
	markdownWg.Add(1)
	go c.markdownWorker(ctx, channels.MarkdownTask, &markdownWg)

	// 啟動內容解析器
	parserCount := c.Config.Crawler.ParserCount
	parsersWg.Add(parserCount)
	for i := 0; i < parserCount; i++ {
		go c.contentParser(ctx, &parsersWg, channels.ArticleInfo, channels.DownloadTask, channels.MarkdownTask)
	}

	return &Workers{
		Parsers:     &parsersWg,
		Downloaders: &downloadersWg,
		Markdown:    &markdownWg,
	}
}

// startProducer 根據模式啟動相應的生產者
func (c *Crawler) startProducer(ctx context.Context, articleChan chan<- types.ArticleInfo) {
	if c.FileURL != "" {
		c.articleProducerFromFile(ctx, articleChan)
	} else {
		c.articleProducer(ctx, articleChan)
	}
}

// waitAndCleanup 等待所有工人完成並進行清理
func (c *Crawler) waitAndCleanup(workers *Workers, channels *WorkerChannels) {
	// 啟動一個 goroutine，等待所有文章解析完成後，關閉下載和 Markdown 任務 channel
	go func() {
		workers.Parsers.Wait()
		close(channels.DownloadTask)
		close(channels.MarkdownTask)
	}()

	// 等待所有下載和 Markdown 任務完成
	workers.Downloaders.Wait()
	workers.Markdown.Wait()
}

// logCompletion 記錄完成信息
func (c *Crawler) logCompletion(ctx context.Context, startTime time.Time) {
	duration := time.Since(startTime)

	if ctx.Err() != nil {
		log.Printf("爬蟲因中斷信號而結束，總耗時: %s", duration)
	} else {
		log.Printf("爬蟲結束，總耗時: %s", duration)
	}

	// 記錄最終記憶體狀態
	if c.Optimizer != nil {
		finalStats := c.Optimizer.GetMemoryStats()
		log.Printf("最終記憶體狀態: %s", finalStats.String())
	}
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

	// 啟動效能監控
	if c.Optimizer != nil {
		c.Optimizer.Start(ctx)
		defer c.Optimizer.Stop()

		// 記錄初始記憶體狀態
		initialStats := c.Optimizer.GetMemoryStats()
		log.Printf("初始記憶體狀態: %s", initialStats.String())
	}

	// 初始化 channels 和 workers
	channels := c.initializeChannels()
	workers := c.startWorkers(ctx, channels)

	// 啟動生產者
	c.startProducer(ctx, channels.ArticleInfo)

	// 等待完成和清理
	c.waitAndCleanup(workers, channels)

	// 記錄完成信息和最終記憶體狀態
	c.logCompletion(ctx, startTime)
}

// articleProducer 產生文章資訊到 channel
func (c *Crawler) articleProducer(ctx context.Context, articleInfoChan chan<- types.ArticleInfo) {
	defer close(articleInfoChan)

	maxPage, err := c.Parser.GetMaxPage(ctx, c.Client, c.Board)
	if err != nil {
		if ctx.Err() != nil {
			log.Printf("獲取最大頁數時被中斷: %v", ctx.Err())
			return
		}
		log.Printf("獲取最大頁數失敗: %v", err)
		return
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
		pageURL := fmt.Sprintf("%s/bbs/%s/index%d.html", constants.PttBaseURL, c.Board, currentPage)
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

		articles, err := c.Parser.ParseArticles(resp.Body)
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
			c.processArticle(ctx, article, downloadTaskChan, markdownTaskChan)
		}
	}
}

// processArticle 處理單一文章的解析和任務分派
func (c *Crawler) processArticle(ctx context.Context, article types.ArticleInfo, downloadTaskChan chan<- types.DownloadTask, markdownTaskChan chan<- types.MarkdownInfo) {
	logMsg := c.getLogMessage(article)
	log.Printf("正在解析文章: %s", logMsg)

	if c.shouldStop(ctx, "內容解析器在延遲時被中斷") {
		return
	}

	parsedTitle, imgURLs, err := c.fetchAndParseArticle(ctx, article)
	if err != nil {
		return // 錯誤已在函數內記錄
	}

	finalTitle := c.determineFinalTitle(article, parsedTitle)

	if len(imgURLs) > 0 {
		c.dispatchTasks(ctx, finalTitle, article, imgURLs, downloadTaskChan, markdownTaskChan)
	}
}

// getLogMessage 獲取用於記錄的消息
func (c *Crawler) getLogMessage(article types.ArticleInfo) string {
	if article.Title != "" {
		return article.Title
	}
	return article.URL
}

// shouldStop 檢查是否應該停止，並處理延遲
func (c *Crawler) shouldStop(ctx context.Context, msg string) bool {
	minDelay, maxDelay := c.Config.GetDelayRange()
	delayRange := int(maxDelay - minDelay)
	delay := minDelay + time.Duration(rand.Intn(delayRange/int(time.Millisecond)))*time.Millisecond

	select {
	case <-ctx.Done():
		log.Println(msg)
		return true
	case <-time.After(delay):
		return false
	}
}

// fetchAndParseArticle 獲取並解析文章內容
func (c *Crawler) fetchAndParseArticle(ctx context.Context, article types.ArticleInfo) (string, []string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", article.URL, nil)
	if err != nil {
		log.Printf("建立文章請求失敗: %s, 錯誤: %v", article.URL, err)
		return "", nil, err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			log.Println("文章爬取被中斷")
			return "", nil, err
		}
		log.Printf("爬取文章頁失敗: %s, 錯誤: %v", article.URL, err)
		return "", nil, err
	}
	defer resp.Body.Close()

	parsedTitle, imgURLs, err := c.Parser.ParseArticleContent(resp.Body)
	if err != nil {
		log.Printf("解析文章頁失敗: %s, 錯誤: %v", article.URL, err)
		return "", nil, err
	}

	return parsedTitle, imgURLs, nil
}

// determineFinalTitle 決定最終使用的標題
func (c *Crawler) determineFinalTitle(article types.ArticleInfo, parsedTitle string) string {
	finalTitle := article.Title
	if (c.FileURL != "" && parsedTitle != "") || (finalTitle == "" && parsedTitle != "") {
		finalTitle = parsedTitle
	}
	return finalTitle
}

// dispatchTasks 分派下載和 Markdown 任務
func (c *Crawler) dispatchTasks(ctx context.Context, finalTitle string, article types.ArticleInfo, imgURLs []string, downloadTaskChan chan<- types.DownloadTask, markdownTaskChan chan<- types.MarkdownInfo) {
	dirName := fmt.Sprintf("%s_%d", cleanFileName(finalTitle), article.PushRate)
	saveDir := filepath.Join(c.Board, dirName)

	// 分派下載任務
	for _, imgURL := range imgURLs {
		if c.dispatchDownloadTask(ctx, imgURL, saveDir, downloadTaskChan) {
			return // 被中斷
		}
	}

	// 分派 Markdown 產生任務
	c.dispatchMarkdownTask(ctx, finalTitle, article, imgURLs, saveDir, markdownTaskChan)
}

// dispatchDownloadTask 分派單個下載任務
func (c *Crawler) dispatchDownloadTask(ctx context.Context, imgURL, saveDir string, downloadTaskChan chan<- types.DownloadTask) bool {
	fileName := c.generateFileName(imgURL)

	select {
	case <-ctx.Done():
		log.Println("分派下載任務時被中斷")
		return true
	case downloadTaskChan <- types.DownloadTask{
		ImageURL: imgURL,
		SavePath: filepath.Join(saveDir, fileName),
	}:
		return false
	}
}

// generateFileName 生成檔案名稱
func (c *Crawler) generateFileName(imgURL string) string {
	fileName := filepath.Base(imgURL)
	if strings.Contains(imgURL, "imgur.com") && !strings.Contains(fileName, ".") {
		if parsedURL, err := url.Parse(imgURL); err == nil {
			fileName = filepath.Base(parsedURL.Path) + ".jpg"
		}
	}
	return fileName
}

// dispatchMarkdownTask 分派 Markdown 任務
func (c *Crawler) dispatchMarkdownTask(ctx context.Context, finalTitle string, article types.ArticleInfo, imgURLs []string, saveDir string, markdownTaskChan chan<- types.MarkdownInfo) {
	select {
	case <-ctx.Done():
		log.Println("分派 Markdown 任務時被中斷")
	case markdownTaskChan <- types.MarkdownInfo{
		Title:      finalTitle,
		ArticleURL: article.URL,
		PushCount:  article.PushRate,
		ImageURLs:  imgURLs,
		SaveDir:    saveDir,
	}:
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
			if err := c.MarkdownGenerator.Generate(task); err != nil {
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
				if err := os.MkdirAll(dir, constants.DirPermission); err != nil {
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
		log.Printf("開啟檔案失敗: %v", err)
		return
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
		log.Printf("讀取檔案時發生錯誤: %v", err)
	}
}
