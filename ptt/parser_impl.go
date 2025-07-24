package ptt

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/twtrubiks/ptt-spider-go/constants"
	"github.com/twtrubiks/ptt-spider-go/errors"
	"github.com/twtrubiks/ptt-spider-go/interfaces"
	"github.com/twtrubiks/ptt-spider-go/types"
)

// ParserImpl 實現 Parser 介面
type ParserImpl struct{}

// NewParser 建立新的解析器實例
func NewParser() interfaces.Parser {
	return &ParserImpl{}
}

// ParseArticles 實現 Parser 介面的 ParseArticles 方法
func (p *ParserImpl) ParseArticles(r io.Reader) ([]types.ArticleInfo, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, errors.NewParseError("建立 goquery 文檔失敗", err)
	}

	var articles []types.ArticleInfo
	doc.Find(".r-ent").Each(func(i int, s *goquery.Selection) {
		titleNode := s.Find(".title a")
		if titleNode.Length() == 0 {
			return // 處理被刪除的文章
		}

		url, _ := titleNode.Attr("href")
		title := strings.TrimSpace(titleNode.Text())

		// 如果標題包含 "公告"，則跳過這篇文章
		if strings.Contains(title, "公告") {
			return
		}

		author := strings.TrimSpace(s.Find(".meta .author").Text())
		pushRateStr := strings.TrimSpace(s.Find(".nrec span").Text())

		pushRate := 0
		if pushRateStr == "爆" {
			pushRate = 100
		} else if strings.HasPrefix(pushRateStr, "X") {
			// 處理 "XX" 和 "X" 開頭的噓文
			rate, err := strconv.Atoi(pushRateStr[1:])
			if err == nil {
				pushRate = -rate
			}
		} else {
			pushRate, _ = strconv.Atoi(pushRateStr)
		}

		articles = append(articles, types.ArticleInfo{
			Title:    title,
			URL:      constants.PttBaseURL + url,
			Author:   author,
			PushRate: pushRate,
		})
	})

	return articles, nil
}

// ParseArticleContent 實現 Parser 介面的 ParseArticleContent 方法
func (p *ParserImpl) ParseArticleContent(r io.Reader) (string, []string, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return "", nil, errors.NewParseError("建立 goquery 文檔失敗", err)
	}

	// 提取文章標題
	title := ""
	doc.Find(".article-meta-tag").EachWithBreak(func(i int, s *goquery.Selection) bool {
		if strings.TrimSpace(s.Text()) == "標題" {
			title = strings.TrimSpace(s.Next().Text())
			return false // 找到後就停止遍歷
		}
		return true
	})

	// 提取圖片 URL
	var imgURLs []string
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			return
		}

		// 簡單的圖片 URL 過濾邏輯
		if strings.HasSuffix(href, ".jpg") || strings.HasSuffix(href, ".jpeg") || strings.HasSuffix(href, ".png") || strings.HasSuffix(href, ".gif") {
			if strings.HasPrefix(href, "//") {
				href = "https:" + href
			} else if strings.HasPrefix(href, "http://") { // 新增：將 http 轉換為 https
				href = "https://" + href[7:]
			}
			imgURLs = append(imgURLs, href)
		} else if strings.Contains(href, "imgur.com/") && !strings.Contains(href, "imgur.com/a/") {
			// 處理沒有副檔名的 imgur 連結
			imgURLs = append(imgURLs, href+".jpg")
		}
	})

	return title, imgURLs, nil
}

// GetMaxPage 實現 Parser 介面的 GetMaxPage 方法
func (p *ParserImpl) GetMaxPage(ctx context.Context, client interfaces.HTTPClient, board string) (int, error) {
	url := fmt.Sprintf("%s/bbs/%s/index.html", constants.PttBaseURL, board)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, errors.NewNetworkError("建立請求失敗", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, errors.NewNetworkError("發送請求失敗", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, errors.NewNetworkError(fmt.Sprintf("HTTP 狀態錯誤: %d", resp.StatusCode), nil)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return 0, errors.NewParseError("解析 HTML 失敗", err)
	}

	prevPageURL, exists := doc.Find(".btn-group-paging a:contains('‹ 上頁')").Attr("href")
	if !exists {
		return 0, errors.NewParseError("無法找到上一頁按鈕", nil)
	}

	// 從 /bbs/Beauty/index2345.html 中提取 2345
	parts := strings.Split(strings.Trim(prevPageURL, ".html"), "index")
	if len(parts) < 2 {
		return 0, errors.NewParseError("無法解析頁碼", nil)
	}

	maxPage, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, errors.NewParseError("頁碼轉換失敗", err)
	}

	return maxPage + 1, nil
}
