// Package main 實作了一個高效能的 PTT 網路爬蟲.
//
// # 概述
//
// 這個爬蟲採用現代化的並行架構，能夠同時處理多篇文章和圖片下載，
// 大幅提升爬取效率。支援兩種操作模式：
//
//   - 看板模式：爬取指定看板的最新文章
//   - 檔案模式：從文字檔讀取文章 URL 列表
//
// # 核心特色
//
//   - 高效並行架構：採用 Goroutine 和 Channel 實現多工處理
//   - 智能圖片處理：自動識別和下載多種圖片格式
//   - 自動化文件生成：為每篇文章生成包含圖片預覽的 Markdown 檔案
//   - 反爬蟲機制：內建延遲機制和模擬瀏覽器行為
//   - Context 優雅關閉：支援 Ctrl+C 中斷信號
//   - 配置管理：支援 YAML 配置檔案，實現參數外部化
//
// # 使用範例
//
//	// 爬取 Beauty 看板最新 5 頁，推文數 >= 20
//	go run main.go -board=beauty -pages=5 -push=20
//
//	// 使用自定義配置檔案
//	go run main.go -config=my-config.yaml -board=beauty -pages=5
//
//	// 從檔案爬取
//	go run main.go -file=urls.txt
//
// # 架構設計
//
// 專案採用生產者-消費者模式，透過 Channel 進行 Goroutine 間的通訊：
//
//  1. Article Producer: 負責產生文章 URL 列表
//  2. Content Parser: 解析文章內容並提取圖片 URL（多個併發）
//  3. Download Worker: 執行圖片下載任務（多個併發）
//  4. Markdown Worker: 生成 Markdown 檔案（單個）
//
// # 配置管理
//
// 支援 YAML 配置檔案，可調整並行度、延遲時間、緩衝區大小等參數：
//
//	crawler:
//	  workers: 10
//	  parserCount: 10
//	  delays:
//	    minMs: 500
//	    maxMs: 2000
//
// # 注意事項
//
// 本工具僅供學習和研究用途，請遵守 PTT 的使用條款和相關法律規定。
// 使用時請適度控制爬取頻率，避免對伺服器造成過大負擔。
package main
