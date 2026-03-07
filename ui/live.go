package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/twtrubiks/ptt-spider-go/types"
)

const (
	maxLogEntries    = 12
	workerStatusIdle = "idle"
)

// LiveConfig 設定即時進度 TUI 的參數
type LiveConfig struct {
	Board    string
	Pages    int
	PushRate int
}

// progressMsg 包裝 ProgressEvent 為 Bubble Tea 訊息
type progressMsg types.ProgressEvent

// progressDoneMsg 表示進度 channel 已關閉
type progressDoneMsg struct{}

// liveModel 是 Bubble Tea 的即時進度 Model
type liveModel struct {
	cfg        LiveConfig
	progressCh <-chan types.ProgressEvent
	cancel     context.CancelFunc
	width      int

	// 進度狀態
	pagesCurrent int
	pagesTotal   int
	articlesOK   int
	imagesTotal  int
	downloadOK   int
	downloadFail int

	// Worker 狀態：workerID → 目前訊息
	workers     map[int]string
	maxWorkerID int

	// 最近事件 log
	logs []string

	// 完成狀態
	done     bool
	doneMsg  string
	quitting bool

	// 進度條
	pageBar     progress.Model
	downloadBar progress.Model
}

func newLiveModel(cfg LiveConfig, progressCh <-chan types.ProgressEvent, cancel context.CancelFunc) liveModel {
	pageBar := progress.New(progress.WithGradient("#5A56E0", "#04B575"), progress.WithWidth(40))
	downloadBar := progress.New(progress.WithGradient("#FF6B6B", "#4ECDC4"), progress.WithWidth(40))

	return liveModel{
		cfg:         cfg,
		progressCh:  progressCh,
		cancel:      cancel,
		width:       80,
		pagesTotal:  cfg.Pages,
		workers:     make(map[int]string),
		pageBar:     pageBar,
		downloadBar: downloadBar,
	}
}

func (m liveModel) Init() tea.Cmd {
	return waitForProgress(m.progressCh)
}

func (m liveModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			m.cancel()
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		barWidth := msg.Width - 30
		if barWidth < 20 {
			barWidth = 20
		}
		if barWidth > 60 {
			barWidth = 60
		}
		m.pageBar.Width = barWidth
		m.downloadBar.Width = barWidth

	case progressMsg:
		evt := types.ProgressEvent(msg)
		m.handleEvent(evt)

		if m.done {
			return m, tea.Sequence(
				tea.Tick(1500*time.Millisecond, func(_ time.Time) tea.Msg {
					return progressDoneMsg{}
				}),
			)
		}
		return m, waitForProgress(m.progressCh)

	case progressDoneMsg:
		return m, tea.Quit
	}

	return m, nil
}

func (m *liveModel) handleEvent(evt types.ProgressEvent) {
	ts := time.Now().Format("15:04:05")

	switch evt.Type {
	case types.EventPageParsed:
		m.pagesCurrent = evt.CurrentPage
		m.pagesTotal = evt.TotalPages
		m.addLog(ts, "INFO", evt.Message)

	case types.EventArticleParsed:
		m.articlesOK++
		m.imagesTotal += evt.ImageCount
		m.addLog(ts, " OK ", fmt.Sprintf("文章「%s」: %d 張圖片", evt.ArticleTitle, evt.ImageCount))

	case types.EventDownloadStart:
		m.workers[evt.WorkerID] = evt.Message
		if evt.WorkerID > m.maxWorkerID {
			m.maxWorkerID = evt.WorkerID
		}

	case types.EventDownloadDone:
		m.downloadOK++
		m.workers[evt.WorkerID] = workerStatusIdle
		m.addLog(ts, " OK ", fmt.Sprintf("Worker#%d 下載完成", evt.WorkerID))

	case types.EventDownloadFail:
		m.downloadFail++
		m.workers[evt.WorkerID] = workerStatusIdle
		m.addLog(ts, " ERR", fmt.Sprintf("Worker#%d %s", evt.WorkerID, evt.Message))

	case types.EventCrawlerDone:
		m.done = true
		m.doneMsg = evt.Message
		m.addLog(ts, "DONE", evt.Message)
	}
}

func (m *liveModel) addLog(ts, level, msg string) {
	entry := fmt.Sprintf("%s [%s] %s", ts, level, msg)
	m.logs = append(m.logs, entry)
	if len(m.logs) > maxLogEntries {
		m.logs = m.logs[len(m.logs)-maxLogEntries:]
	}
}

func (m liveModel) View() string {
	if m.quitting {
		return "正在停止爬蟲...\n"
	}

	var b strings.Builder

	// 標題
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	b.WriteString(titleStyle.Render("PTT Spider"))
	b.WriteString("\n")

	// 設定資訊
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	b.WriteString(dimStyle.Render(fmt.Sprintf("看板: %s    頁數: %d    推文門檻: ≥%d",
		m.cfg.Board, m.cfg.Pages, m.cfg.PushRate)))
	b.WriteString("\n\n")

	// 頁面進度
	pagePercent := 0.0
	if m.pagesTotal > 0 {
		pagePercent = float64(m.pagesCurrent) / float64(m.pagesTotal)
	}
	fmt.Fprintf(&b, "  頁面  %s %d/%d",
		m.pageBar.ViewAs(pagePercent), m.pagesCurrent, m.pagesTotal)
	b.WriteString("\n")

	// 下載進度
	downloadTotal := m.imagesTotal
	downloadDone := m.downloadOK + m.downloadFail
	downloadPercent := 0.0
	if downloadTotal > 0 {
		downloadPercent = float64(downloadDone) / float64(downloadTotal)
	}
	downloadInfo := fmt.Sprintf("%d/%d", downloadDone, downloadTotal)
	if m.downloadFail > 0 {
		downloadInfo += fmt.Sprintf(" (%d 失敗)", m.downloadFail)
	}
	fmt.Fprintf(&b, "  下載  %s %s",
		m.downloadBar.ViewAs(downloadPercent), downloadInfo)
	b.WriteString("\n")

	// 文章統計
	b.WriteString(dimStyle.Render(fmt.Sprintf("  文章: %d 篇已解析，共發現 %d 張圖片",
		m.articlesOK, m.imagesTotal)))
	b.WriteString("\n\n")

	// Worker 狀態
	if m.maxWorkerID > 0 {
		workerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
		activeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
		idleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("242"))

		for id := 1; id <= m.maxWorkerID; id++ {
			status, ok := m.workers[id]
			if !ok {
				continue
			}
			if status == workerStatusIdle || status == "" {
				b.WriteString(workerStyle.Render(fmt.Sprintf("  Worker#%d ", id)))
				b.WriteString(idleStyle.Render(workerStatusIdle))
			} else {
				b.WriteString(workerStyle.Render(fmt.Sprintf("  Worker#%d ", id)))
				b.WriteString(activeStyle.Render(truncate(status, 50)))
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// 最近事件 log
	if len(m.logs) > 0 {
		logHeaderStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("252"))
		logStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

		b.WriteString(logHeaderStyle.Render("最近事件"))
		b.WriteString("\n")
		for _, entry := range m.logs {
			b.WriteString(logStyle.Render("  " + entry))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// 底部
	if m.done {
		doneStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
		b.WriteString(doneStyle.Render(fmt.Sprintf("完成！%s", m.doneMsg)))
		b.WriteString("\n")
	} else {
		b.WriteString(dimStyle.Render("Ctrl+C 優雅停止"))
		b.WriteString("\n")
	}

	return b.String()
}

// waitForProgress 建立一個 tea.Cmd 等待下一個進度事件
func waitForProgress(ch <-chan types.ProgressEvent) tea.Cmd {
	return func() tea.Msg {
		evt, ok := <-ch
		if !ok {
			return progressDoneMsg{}
		}
		return progressMsg(evt)
	}
}

// truncate 截斷字串到指定長度
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// RunLiveTUI 啟動即時進度 TUI，阻塞直到爬蟲完成或使用者中斷。
// progressCh 由 Crawler 的 WithProgress option 注入的同一個 channel。
// cancel 用於在使用者按 Ctrl+C 時取消 crawler 的 context。
func RunLiveTUI(cfg LiveConfig, progressCh <-chan types.ProgressEvent, cancel context.CancelFunc) error {
	m := newLiveModel(cfg, progressCh, cancel)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
