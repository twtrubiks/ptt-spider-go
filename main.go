package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/twtrubiks/ptt-spider-go/config"
	"github.com/twtrubiks/ptt-spider-go/constants"
	"github.com/twtrubiks/ptt-spider-go/crawler"
)

func main() {
	// 定義命令列參數
	board := flag.String("board", constants.DefaultBoard, "看板名稱")
	pages := flag.Int("pages", constants.DefaultPages, "要爬取的頁數")
	pushRate := flag.Int("push", constants.DefaultPushRate, "推文數門檻")
	fileURL := flag.String("file", "", "包含文章 URL 的文字檔路徑 (優先於看板模式)")
	configPath := flag.String("config", "config.yaml", "配置檔案路徑")

	flag.Parse()

	// 載入配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("載入配置失敗: %v", err)
	}

	// 建立爬蟲
	c, err := crawler.NewCrawler(*board, *pages, *pushRate, *fileURL, cfg)
	if err != nil {
		log.Fatalf("建立爬蟲失敗: %v", err)
	}

	// 建立 context 並監聽中斷信號
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 監聽系統信號
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("收到中斷信號，正在優雅關閉爬蟲...")
		cancel()
	}()

	// 啟動爬蟲
	c.Run(ctx)
}
