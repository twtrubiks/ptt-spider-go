package crawler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/twtrubiks/ptt-spider-go/constants"
	"github.com/twtrubiks/ptt-spider-go/mocks"
)

func TestDoWithRetry_ImmediateSuccess(t *testing.T) {
	client := &mocks.MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("ok")),
			}, nil
		},
	}

	req, _ := http.NewRequestWithContext(context.Background(), "GET", "http://example.com", nil)
	resp, err := doWithRetry(context.Background(), client, req)

	if err != nil {
		t.Fatalf("期望無錯誤，但收到: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("期望狀態碼 200，但收到: %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestDoWithRetry_429ThenSuccess(t *testing.T) {
	callCount := 0
	client := &mocks.MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			callCount++
			if callCount == 1 {
				return &http.Response{
					StatusCode: http.StatusTooManyRequests,
					Header:     http.Header{"Retry-After": []string{"0"}},
					Body:       io.NopCloser(strings.NewReader("")),
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("ok")),
			}, nil
		},
	}

	req, _ := http.NewRequestWithContext(context.Background(), "GET", "http://example.com", nil)
	resp, err := doWithRetry(context.Background(), client, req)

	if err != nil {
		t.Fatalf("期望無錯誤，但收到: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("期望狀態碼 200，但收到: %d", resp.StatusCode)
	}
	if callCount != 2 {
		t.Fatalf("期望呼叫 2 次，但實際呼叫 %d 次", callCount)
	}
	resp.Body.Close()
}

func TestDoWithRetry_ExhaustedRetries(t *testing.T) {
	callCount := 0
	client := &mocks.MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			callCount++
			return &http.Response{
				StatusCode: http.StatusTooManyRequests,
				Header:     http.Header{"Retry-After": []string{"0"}},
				Body:       io.NopCloser(strings.NewReader("")),
			}, nil
		},
	}

	req, _ := http.NewRequestWithContext(context.Background(), "GET", "http://example.com", nil)
	resp, err := doWithRetry(context.Background(), client, req)

	if err == nil {
		t.Fatal("期望收到錯誤，但沒有")
	}
	if resp != nil {
		t.Fatal("期望 resp 為 nil")
	}
	if callCount != constants.RetryMaxAttempts {
		t.Fatalf("期望呼叫 %d 次，但實際呼叫 %d 次", constants.RetryMaxAttempts, callCount)
	}
	if !strings.Contains(err.Error(), "429") {
		t.Fatalf("錯誤訊息應包含 429，但收到: %v", err)
	}
}

func TestDoWithRetry_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	callCount := 0
	client := &mocks.MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			callCount++
			if callCount == 1 {
				// 第一次回傳 429 後取消 context
				cancel()
				return &http.Response{
					StatusCode: http.StatusTooManyRequests,
					Header:     http.Header{},
					Body:       io.NopCloser(strings.NewReader("")),
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("ok")),
			}, nil
		},
	}

	req, _ := http.NewRequestWithContext(ctx, "GET", "http://example.com", nil)
	resp, err := doWithRetry(ctx, client, req)

	if err == nil {
		t.Fatal("期望收到錯誤，但沒有")
	}
	if resp != nil {
		t.Fatal("期望 resp 為 nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("期望 context.Canceled 錯誤，但收到: %v", err)
	}
}

func TestDoWithRetry_NetworkError(t *testing.T) {
	networkErr := errors.New("connection refused")
	client := &mocks.MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return nil, networkErr
		},
	}

	req, _ := http.NewRequestWithContext(context.Background(), "GET", "http://example.com", nil)
	resp, err := doWithRetry(context.Background(), client, req)

	if err == nil {
		t.Fatal("期望收到錯誤，但沒有")
	}
	if resp != nil {
		t.Fatal("期望 resp 為 nil")
	}
	if !errors.Is(err, networkErr) {
		t.Fatalf("期望網路錯誤，但收到: %v", err)
	}
}

func TestDoWithRetry_Non429ErrorCode(t *testing.T) {
	client := &mocks.MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader("not found")),
			}, nil
		},
	}

	req, _ := http.NewRequestWithContext(context.Background(), "GET", "http://example.com", nil)
	resp, err := doWithRetry(context.Background(), client, req)

	if err != nil {
		t.Fatalf("期望無錯誤，但收到: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("期望狀態碼 404，但收到: %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestDoWithRetry_RetryAfterHeaderSeconds(t *testing.T) {
	callCount := 0
	client := &mocks.MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			callCount++
			if callCount == 1 {
				return &http.Response{
					StatusCode: http.StatusTooManyRequests,
					Header:     http.Header{"Retry-After": []string{"1"}},
					Body:       io.NopCloser(strings.NewReader("")),
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("ok")),
			}, nil
		},
	}

	req, _ := http.NewRequestWithContext(context.Background(), "GET", "http://example.com", nil)
	start := time.Now()
	resp, err := doWithRetry(context.Background(), client, req)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("期望無錯誤，但收到: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("期望狀態碼 200，但收到: %d", resp.StatusCode)
	}
	// 應至少等待約 1 秒
	if elapsed < 900*time.Millisecond {
		t.Fatalf("期望至少等待 ~1 秒，但只等了 %v", elapsed)
	}
	resp.Body.Close()
}

func TestCalcRetryDelay_ExponentialBackoff(t *testing.T) {
	tests := []struct {
		name     string
		attempt  int
		expected time.Duration
	}{
		{"第 1 次", 1, time.Duration(constants.RetryInitialDelayMs) * time.Millisecond},
		{"第 2 次", 2, time.Duration(constants.RetryInitialDelayMs*constants.RetryBackoffFactor) * time.Millisecond},
		{"第 3 次", 3, time.Duration(constants.RetryInitialDelayMs*constants.RetryBackoffFactor*constants.RetryBackoffFactor) * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				Header: http.Header{},
				Body:   io.NopCloser(strings.NewReader("")),
			}
			delay := calcRetryDelay(resp, tt.attempt)
			if delay != tt.expected {
				t.Errorf("期望延遲 %v，但收到 %v", tt.expected, delay)
			}
		})
	}
}

func TestCalcRetryDelay_MaxDelayCap(t *testing.T) {
	resp := &http.Response{
		Header: http.Header{},
		Body:   io.NopCloser(strings.NewReader("")),
	}
	// 使用很大的 attempt 值觸發上限
	delay := calcRetryDelay(resp, 100)
	maxDelay := time.Duration(constants.RetryMaxDelayMs) * time.Millisecond
	if delay != maxDelay {
		t.Errorf("期望延遲上限 %v，但收到 %v", maxDelay, delay)
	}
}

func TestCalcRetryDelay_RetryAfterSeconds(t *testing.T) {
	resp := &http.Response{
		Header: http.Header{"Retry-After": []string{"5"}},
		Body:   io.NopCloser(strings.NewReader("")),
	}
	delay := calcRetryDelay(resp, 1)
	expected := 5 * time.Second
	if delay != expected {
		t.Errorf("期望延遲 %v，但收到 %v", expected, delay)
	}
}

func TestCalcRetryDelay_RetryAfterSecondsCapped(t *testing.T) {
	resp := &http.Response{
		Header: http.Header{"Retry-After": []string{"60"}},
		Body:   io.NopCloser(strings.NewReader("")),
	}
	delay := calcRetryDelay(resp, 1)
	maxDelay := time.Duration(constants.RetryMaxDelayMs) * time.Millisecond
	if delay != maxDelay {
		t.Errorf("期望延遲上限 %v，但收到 %v", maxDelay, delay)
	}
}

func TestCalcRetryDelay_NilResponse(t *testing.T) {
	delay := calcRetryDelay(nil, 1)
	expected := time.Duration(constants.RetryInitialDelayMs) * time.Millisecond
	if delay != expected {
		t.Errorf("期望延遲 %v，但收到 %v", expected, delay)
	}
}
