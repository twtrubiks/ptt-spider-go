package crawler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/twtrubiks/ptt-spider-go/constants"
	"github.com/twtrubiks/ptt-spider-go/interfaces"
	"github.com/twtrubiks/ptt-spider-go/internal/ioutil"
)

// doWithRetry 包裝 client.Do，收到 HTTP 429 時自動以指數退避重試。
// 非 429 錯誤碼或網路錯誤不會重試，直接回傳。
// 重試用盡後回傳 nil 和錯誤。
func doWithRetry(ctx context.Context, client interfaces.HTTPClient, req *http.Request) (*http.Response, error) {
	for attempt := 1; attempt <= constants.RetryMaxAttempts; attempt++ {
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusTooManyRequests {
			return resp, nil
		}

		// 429: 計算退避時間並重試
		delay := calcRetryDelay(resp, attempt)
		log.Printf("收到 HTTP 429，第 %d/%d 次重試，等待 %v: %s", attempt, constants.RetryMaxAttempts, delay, req.URL)

		ioutil.CloseWithLog(resp.Body, "429 重試回應 Body")

		if attempt == constants.RetryMaxAttempts {
			return nil, fmt.Errorf("重試 %d 次後仍收到 429: %s", constants.RetryMaxAttempts, req.URL)
		}

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}

	// 不應到達此處，但作為安全網
	return nil, fmt.Errorf("重試邏輯異常: %s", req.URL)
}

// calcRetryDelay 計算第 attempt 次重試的等待時間。
// 若 response 包含 Retry-After header（秒數或 HTTP-date 格式），則優先使用。
// 否則使用指數退避公式：initialDelay * factor^(attempt-1)，上限為 maxDelay。
func calcRetryDelay(resp *http.Response, attempt int) time.Duration {
	if resp != nil {
		if ra := resp.Header.Get("Retry-After"); ra != "" {
			// 嘗試解析為秒數
			if seconds, err := strconv.Atoi(ra); err == nil && seconds > 0 {
				delay := time.Duration(seconds) * time.Second
				maxDelay := time.Duration(constants.RetryMaxDelayMs) * time.Millisecond
				if delay > maxDelay {
					return maxDelay
				}
				return delay
			}
			// 嘗試解析為 HTTP-date 格式
			if t, err := http.ParseTime(ra); err == nil {
				delay := time.Until(t)
				if delay <= 0 {
					delay = time.Duration(constants.RetryInitialDelayMs) * time.Millisecond
				}
				maxDelay := time.Duration(constants.RetryMaxDelayMs) * time.Millisecond
				if delay > maxDelay {
					return maxDelay
				}
				return delay
			}
		}
	}

	// 指數退避: initialDelay * factor^(attempt-1)
	delay := constants.RetryInitialDelayMs
	for i := 1; i < attempt; i++ {
		delay *= constants.RetryBackoffFactor
		if delay >= constants.RetryMaxDelayMs {
			return time.Duration(constants.RetryMaxDelayMs) * time.Millisecond
		}
	}
	return time.Duration(delay) * time.Millisecond
}
