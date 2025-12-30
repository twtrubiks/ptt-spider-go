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
	"github.com/twtrubiks/ptt-spider-go/types"
)

// ParseArticles 從看板列表頁解析文章資訊
func ParseArticles(r io.Reader) ([]types.ArticleInfo, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
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
		switch {
		case pushRateStr == "爆":
			pushRate = 100
		case strings.HasPrefix(pushRateStr, "X"):
			// 處理 "XX" 和 "X" 開頭的噓文
			rate, err := strconv.Atoi(pushRateStr[1:])
			if err == nil {
				pushRate = -rate
			}
		default:
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

// ParseArticleContent 從文章頁面解析標題和圖片 URL
func ParseArticleContent(r io.Reader) (string, []string, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return "", nil, err
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

// GetMaxPage 獲取最大頁數
func GetMaxPage(ctx context.Context, client *http.Client, board string) (int, error) {
	url := fmt.Sprintf("%s/bbs/%s/index.html", constants.PttBaseURL, board)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return 0, err
	}

	prevPageURL, exists := doc.Find(".btn-group-paging a:contains('‹ 上頁')").Attr("href")
	if !exists {
		return 0, fmt.Errorf("無法找到上一頁按鈕")
	}

	// 從 /bbs/Beauty/index2345.html 中提取 2345
	parts := strings.Split(strings.Trim(prevPageURL, ".html"), "index")
	if len(parts) < 2 {
		return 0, fmt.Errorf("無法解析頁碼")
	}

	maxPage, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}

	return maxPage + 1, nil
}
