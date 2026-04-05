package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/twtrubiks/ptt-spider-go/config"
	"github.com/twtrubiks/ptt-spider-go/constants"
	"github.com/twtrubiks/ptt-spider-go/crawler"
	"github.com/twtrubiks/ptt-spider-go/types"
	"github.com/twtrubiks/ptt-spider-go/ui"
)

func main() {
	// 定義命令列參數
	board := flag.String("board", constants.DefaultBoard, "看板名稱")
	pages := flag.Int("pages", constants.DefaultPages, "要爬取的頁數")
	pushRate := flag.Int("push", constants.DefaultPushRate, "推文數門檻")
	fileURL := flag.String("file", "", "包含文章 URL 的文字檔路徑 (優先於看板模式)")
	configPath := flag.String("config", "config.yaml", "配置檔案路徑")
	tuiMode := flag.Bool("tui", false, "啟動互動式 TUI 選單（含即時進度畫面）")

	flag.Parse()

	logger := ui.NewStyledLogger()

	// TUI 互動模式
	if *tuiMode {
		tuiCfg, err := ui.RunStartupForm(
			constants.DefaultBoard,
			constants.DefaultPages,
			constants.DefaultPushRate,
		)
		if err != nil {
			logger.Error("TUI 表單錯誤: %v", err)
			os.Exit(1)
		}
		*board = tuiCfg.Board
		*pages = tuiCfg.Pages
		*pushRate = tuiCfg.PushRate
		*fileURL = tuiCfg.FileURL
	}

	// 載入配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Error("載入配置失敗: %v", err)
		os.Exit(1)
	}

	// 建立 context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if *tuiMode {
		runWithTUI(ctx, cancel, logger, *board, *pages, *pushRate, *fileURL, cfg)
	} else {
		runWithCLI(ctx, cancel, logger, *board, *pages, *pushRate, *fileURL, cfg)
	}
}

// runWithTUI 使用即時進度 TUI 模式執行爬蟲
func runWithTUI(ctx context.Context, cancel context.CancelFunc, logger ui.Logger, board string, pages, pushRate int, fileURL string, cfg *config.Config) {
	progressCh := make(chan types.ProgressEvent, 200)

	c, err := crawler.NewCrawler(board, pages, pushRate, fileURL, cfg,
		crawler.WithProgress(progressCh),
		crawler.WithLogger(ui.NewNoopLogger()),
	)
	if err != nil {
		logger.Error("建立爬蟲失敗: %v", err)
		os.Exit(1)
	}

	// 在背景執行爬蟲，完成後關閉 progress channel
	go func() {
		c.Run(ctx)
		close(progressCh)
	}()

	liveCfg := ui.LiveConfig{
		Board:    board,
		Pages:    pages,
		PushRate: pushRate,
	}

	if err := ui.RunLiveTUI(liveCfg, progressCh, cancel); err != nil {
		logger.Error("TUI 錯誤: %v", err)
		os.Exit(1)
	}
}

// runWithCLI 使用傳統 CLI 模式（彩色 log 輸出）執行爬蟲
func runWithCLI(ctx context.Context, cancel context.CancelFunc, logger ui.Logger, board string, pages, pushRate int, fileURL string, cfg *config.Config) {
	c, err := crawler.NewCrawler(board, pages, pushRate, fileURL, cfg)
	if err != nil {
		logger.Error("建立爬蟲失敗: %v", err)
		os.Exit(1)
	}

	// 監聽系統信號
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Warn("收到中斷信號，正在優雅關閉爬蟲...")
		cancel()
	}()

	// 啟動爬蟲
	c.Run(ctx)
}
