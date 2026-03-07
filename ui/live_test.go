package ui

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/twtrubiks/ptt-spider-go/types"
)

func TestNoopLogger_ImplementsInterface(_ *testing.T) {
	var _ Logger = (*NoopLogger)(nil)
	var _ Logger = NewNoopLogger()
}

func TestNoopLogger_NoPanic(t *testing.T) {
	logger := NewNoopLogger()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("NoopLogger panicked: %v", r)
		}
	}()

	logger.Info("test %s", "info")
	logger.Success("test %d", 1)
	logger.Error("test error")
	logger.Warn("test %s %s", "a", "b")
}

func TestNewLiveModel_Defaults(t *testing.T) {
	ch := make(chan types.ProgressEvent, 1)
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := LiveConfig{Board: "beauty", Pages: 5, PushRate: 10}
	m := newLiveModel(cfg, ch, cancel)

	if m.cfg.Board != "beauty" {
		t.Errorf("Board = %q, want %q", m.cfg.Board, "beauty")
	}
	if m.pagesTotal != 5 {
		t.Errorf("pagesTotal = %d, want 5", m.pagesTotal)
	}
	if m.done {
		t.Error("should not be done initially")
	}
	if len(m.logs) != 0 {
		t.Error("logs should be empty initially")
	}
}

func TestLiveModel_HandlePageEvent(t *testing.T) {
	ch := make(chan types.ProgressEvent, 1)
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	m := newLiveModel(LiveConfig{Board: "test", Pages: 3}, ch, cancel)

	evt := types.ProgressEvent{
		Type:        types.EventPageParsed,
		CurrentPage: 2,
		TotalPages:  3,
		Message:     "解析第 2/3 頁完成",
	}
	m.handleEvent(evt)

	if m.pagesCurrent != 2 {
		t.Errorf("pagesCurrent = %d, want 2", m.pagesCurrent)
	}
	if m.pagesTotal != 3 {
		t.Errorf("pagesTotal = %d, want 3", m.pagesTotal)
	}
	if len(m.logs) != 1 {
		t.Errorf("logs count = %d, want 1", len(m.logs))
	}
}

func TestLiveModel_HandleArticleEvent(t *testing.T) {
	ch := make(chan types.ProgressEvent, 1)
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	m := newLiveModel(LiveConfig{Board: "test", Pages: 1}, ch, cancel)

	m.handleEvent(types.ProgressEvent{
		Type:         types.EventArticleParsed,
		ArticleTitle: "Test Article",
		ImageCount:   5,
	})

	if m.articlesOK != 1 {
		t.Errorf("articlesOK = %d, want 1", m.articlesOK)
	}
	if m.imagesTotal != 5 {
		t.Errorf("imagesTotal = %d, want 5", m.imagesTotal)
	}
}

func TestLiveModel_HandleDownloadEvents(t *testing.T) {
	ch := make(chan types.ProgressEvent, 1)
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	m := newLiveModel(LiveConfig{Board: "test", Pages: 1}, ch, cancel)
	m.imagesTotal = 3

	// Download start
	m.handleEvent(types.ProgressEvent{
		Type:     types.EventDownloadStart,
		WorkerID: 1,
		Message:  "http://example.com/img.jpg",
	})

	if m.workers[1] != "http://example.com/img.jpg" {
		t.Errorf("worker 1 status = %q, want URL", m.workers[1])
	}

	// Download done
	m.handleEvent(types.ProgressEvent{
		Type:     types.EventDownloadDone,
		WorkerID: 1,
		Message:  "/path/to/img.jpg",
	})

	if m.downloadOK != 1 {
		t.Errorf("downloadOK = %d, want 1", m.downloadOK)
	}
	if m.workers[1] != workerStatusIdle {
		t.Errorf("worker 1 should be idle after download, got %q", m.workers[1])
	}

	// Download fail
	m.handleEvent(types.ProgressEvent{
		Type:     types.EventDownloadFail,
		WorkerID: 2,
		Message:  "狀態碼 404",
	})

	if m.downloadFail != 1 {
		t.Errorf("downloadFail = %d, want 1", m.downloadFail)
	}
}

func TestLiveModel_HandleCrawlerDone(t *testing.T) {
	ch := make(chan types.ProgressEvent, 1)
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	m := newLiveModel(LiveConfig{Board: "test", Pages: 1}, ch, cancel)

	m.handleEvent(types.ProgressEvent{
		Type:    types.EventCrawlerDone,
		Message: "總耗時: 10s",
	})

	if !m.done {
		t.Error("should be done after CrawlerDone event")
	}
	if m.doneMsg != "總耗時: 10s" {
		t.Errorf("doneMsg = %q, want %q", m.doneMsg, "總耗時: 10s")
	}
}

func TestLiveModel_LogRotation(t *testing.T) {
	ch := make(chan types.ProgressEvent, 1)
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	m := newLiveModel(LiveConfig{Board: "test", Pages: 1}, ch, cancel)

	// 加入超過 maxLogEntries 的 log
	for i := 0; i < maxLogEntries+5; i++ {
		m.handleEvent(types.ProgressEvent{
			Type:         types.EventArticleParsed,
			ArticleTitle: "Article",
			ImageCount:   1,
		})
	}

	if len(m.logs) != maxLogEntries {
		t.Errorf("logs count = %d, want %d (should be capped)", len(m.logs), maxLogEntries)
	}
}

func TestLiveModel_ViewContainsBoard(t *testing.T) {
	ch := make(chan types.ProgressEvent, 1)
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	m := newLiveModel(LiveConfig{Board: "beauty", Pages: 3, PushRate: 10}, ch, cancel)
	view := m.View()

	if !strings.Contains(view, "beauty") {
		t.Error("View should contain board name")
	}
	if !strings.Contains(view, "PTT Spider") {
		t.Error("View should contain title")
	}
	if !strings.Contains(view, "Ctrl+C") {
		t.Error("View should contain quit hint")
	}
}

func TestLiveModel_ViewDone(t *testing.T) {
	ch := make(chan types.ProgressEvent, 1)
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	m := newLiveModel(LiveConfig{Board: "test", Pages: 1}, ch, cancel)
	m.done = true
	m.doneMsg = "總耗時: 5s"

	view := m.View()
	if !strings.Contains(view, "完成") {
		t.Error("done View should contain completion message")
	}
	if !strings.Contains(view, "按 Enter 或 q 離開") {
		t.Error("done View should contain exit hint")
	}
}

func TestLiveModel_DoneEnterQuits(t *testing.T) {
	ch := make(chan types.ProgressEvent, 1)
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	m := newLiveModel(LiveConfig{Board: "test", Pages: 1}, ch, cancel)
	m.done = true

	// Enter should quit when done
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Error("Enter should return quit command when done")
	}
}

func TestLiveModel_EnterIgnoredWhenNotDone(t *testing.T) {
	ch := make(chan types.ProgressEvent, 1)
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	m := newLiveModel(LiveConfig{Board: "test", Pages: 1}, ch, cancel)

	// Enter should be ignored when not done
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("Enter should not return command when not done")
	}
}

func TestLiveModel_UpdateQuit(t *testing.T) {
	ch := make(chan types.ProgressEvent, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m := newLiveModel(LiveConfig{Board: "test", Pages: 1}, ch, cancel)

	// Simulate Ctrl+C
	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	updated := newModel.(liveModel)

	if !updated.quitting {
		t.Error("should be quitting after Ctrl+C")
	}
	if cmd == nil {
		t.Error("should return quit command")
	}
	// context should be cancelled
	if ctx.Err() == nil {
		t.Error("context should be cancelled after Ctrl+C")
	}
}

func TestLiveModel_UpdateProgressMsg(t *testing.T) {
	ch := make(chan types.ProgressEvent, 1)
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	m := newLiveModel(LiveConfig{Board: "test", Pages: 3}, ch, cancel)

	msg := progressMsg(types.ProgressEvent{
		Type:        types.EventPageParsed,
		CurrentPage: 1,
		TotalPages:  3,
		Message:     "test",
	})

	newModel, cmd := m.Update(msg)
	updated := newModel.(liveModel)

	if updated.pagesCurrent != 1 {
		t.Errorf("pagesCurrent = %d, want 1", updated.pagesCurrent)
	}
	if cmd == nil {
		t.Error("should return cmd to wait for next event")
	}
}

func TestWaitForProgress_ChannelClosed(t *testing.T) {
	ch := make(chan types.ProgressEvent)
	close(ch)

	cmd := waitForProgress(ch)
	msg := cmd()

	if _, ok := msg.(progressDoneMsg); !ok {
		t.Errorf("expected progressDoneMsg, got %T", msg)
	}
}

func TestWaitForProgress_ReceivesEvent(t *testing.T) {
	ch := make(chan types.ProgressEvent, 1)
	ch <- types.ProgressEvent{Type: types.EventPageParsed, Message: "test"}

	cmd := waitForProgress(ch)
	msg := cmd()

	pmsg, ok := msg.(progressMsg)
	if !ok {
		t.Fatalf("expected progressMsg, got %T", msg)
	}
	if types.ProgressEvent(pmsg).Type != types.EventPageParsed {
		t.Errorf("got event type %d, want EventPageParsed", types.ProgressEvent(pmsg).Type)
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"this is a longer string", 10, "this is..."},
		{"abc", 3, "abc"},
		{"abcd", 3, "..."},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := truncate(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}
