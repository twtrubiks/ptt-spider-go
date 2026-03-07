package ui

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/charmbracelet/huh"
)

const (
	modeBoard = "board"
	modeFile  = "file"
)

// StartupConfig 儲存 TUI 表單收集到的啟動參數
type StartupConfig struct {
	Board    string
	Pages    int
	PushRate int
	FileURL  string
}

// RunStartupForm 執行互動式 TUI 啟動表單，收集爬蟲參數。
// defaultBoard/defaultPages/defaultPushRate 為使用者未輸入時的預設值。
func RunStartupForm(defaultBoard string, defaultPages, defaultPushRate int) (*StartupConfig, error) {
	var mode string

	modeForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("選擇爬取模式").
				Description("使用方向鍵選擇，Enter 確認").
				Options(
					huh.NewOption("看板模式 — 依看板名稱爬取", modeBoard),
					huh.NewOption("檔案模式 — 從檔案讀取 URL", modeFile),
				).
				Value(&mode),
		),
	)

	if err := modeForm.Run(); err != nil {
		return nil, fmt.Errorf("模式選擇失敗: %w", err)
	}

	if mode == modeFile {
		return runFileForm(defaultBoard)
	}
	return runBoardForm(defaultBoard, defaultPages, defaultPushRate)
}

func runBoardForm(defaultBoard string, defaultPages, defaultPushRate int) (*StartupConfig, error) {
	var board, pagesStr, pushRateStr string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("看板名稱").
				Placeholder(defaultBoard).
				Value(&board),

			huh.NewInput().
				Title("爬取頁數").
				Placeholder(strconv.Itoa(defaultPages)).
				Validate(validatePositiveInt).
				Value(&pagesStr),

			huh.NewInput().
				Title("推文數門檻").
				Placeholder(strconv.Itoa(defaultPushRate)).
				Validate(validateNonNegativeInt).
				Value(&pushRateStr),
		),
	)

	if err := form.Run(); err != nil {
		return nil, fmt.Errorf("參數設定失敗: %w", err)
	}

	return applyBoardDefaults(board, pagesStr, pushRateStr, defaultBoard, defaultPages, defaultPushRate), nil
}

func runFileForm(defaultBoard string) (*StartupConfig, error) {
	var fileURL, board string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("文章 URL 檔案路徑").
				Placeholder("urls.txt").
				Validate(validateNonEmpty).
				Value(&fileURL),

			huh.NewInput().
				Title("看板名稱 (用於存檔目錄)").
				Placeholder(defaultBoard).
				Value(&board),
		),
	)

	if err := form.Run(); err != nil {
		return nil, fmt.Errorf("參數設定失敗: %w", err)
	}

	return applyFileDefaults(fileURL, board, defaultBoard), nil
}

// applyBoardDefaults 將使用者輸入套用預設值，回傳 StartupConfig
func applyBoardDefaults(board, pagesStr, pushRateStr, defaultBoard string, defaultPages, defaultPushRate int) *StartupConfig {
	if board == "" {
		board = defaultBoard
	}

	pages := defaultPages
	if pagesStr != "" {
		if n, err := strconv.Atoi(pagesStr); err == nil {
			pages = n
		}
	}

	pushRate := defaultPushRate
	if pushRateStr != "" {
		if n, err := strconv.Atoi(pushRateStr); err == nil {
			pushRate = n
		}
	}

	return &StartupConfig{
		Board:    board,
		Pages:    pages,
		PushRate: pushRate,
	}
}

// applyFileDefaults 將使用者輸入套用預設值，回傳 StartupConfig
func applyFileDefaults(fileURL, board, defaultBoard string) *StartupConfig {
	if board == "" {
		board = defaultBoard
	}

	return &StartupConfig{
		Board:   board,
		FileURL: fileURL,
	}
}

func validateNonEmpty(s string) error {
	if s == "" {
		return errors.New("此欄位不可為空")
	}
	return nil
}

func validatePositiveInt(s string) error {
	if s == "" {
		return nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return errors.New("請輸入有效數字")
	}
	if n <= 0 {
		return errors.New("數字必須大於 0")
	}
	return nil
}

func validateNonNegativeInt(s string) error {
	if s == "" {
		return nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return errors.New("請輸入有效數字")
	}
	if n < 0 {
		return errors.New("數字不可為負數")
	}
	return nil
}
