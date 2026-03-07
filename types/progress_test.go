package types

import "testing"

func TestEventType_Constants(t *testing.T) {
	// 確認每個事件類型有不同的值
	types := []EventType{
		EventPageParsed,
		EventArticleParsed,
		EventDownloadStart,
		EventDownloadDone,
		EventDownloadFail,
		EventCrawlerDone,
	}

	seen := make(map[EventType]bool)
	for _, et := range types {
		if seen[et] {
			t.Errorf("duplicate EventType value: %d", et)
		}
		seen[et] = true
	}

	if len(seen) != 6 {
		t.Errorf("expected 6 unique EventType values, got %d", len(seen))
	}
}

func TestProgressEvent_ZeroValue(t *testing.T) {
	var evt ProgressEvent

	if evt.Type != EventPageParsed {
		t.Errorf("zero value Type = %d, want EventPageParsed (0)", evt.Type)
	}
	if evt.WorkerID != 0 {
		t.Error("zero value WorkerID should be 0")
	}
	if evt.Message != "" {
		t.Error("zero value Message should be empty")
	}
	if evt.ArticleTitle != "" {
		t.Error("zero value ArticleTitle should be empty")
	}
	if evt.ImageCount != 0 {
		t.Error("zero value ImageCount should be 0")
	}
}
