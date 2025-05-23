package internal

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

type ScanSession struct {
	pw      *playwright.Playwright
	browser playwright.Browser
	context playwright.BrowserContext
}

func InitScanSession() (*ScanSession, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("Playwright 启动失败: %w", err)
	}

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		pw.Stop()
		return nil, fmt.Errorf("浏览器启动失败: %w", err)
	}

	context, err := browser.NewContext()
	if err != nil {
		browser.Close()
		pw.Stop()
		return nil, err
	}

	return &ScanSession{
		pw:      pw,
		browser: browser,
		context: context,
	}, nil
}

func (s *ScanSession) Close() {
	s.context.Close()
	s.browser.Close()
	s.pw.Stop()
}

func (s *ScanSession) RunScan(targetURL string, maxDepth int) ([]ScanResult, error) {
	visited := make(map[string]bool)
	var results []ScanResult

	var scanRecursive func(url string, parent string, currentDepth int) error
	scanRecursive = func(url string, parent string, currentDepth int) error {
		if visited[url] || currentDepth > maxDepth {
			return nil
		}
		visited[url] = true

		fmt.Printf("🔍 扫描 %s (深度 %d)\n", url, currentDepth)
		result, links, err := s.scanSingle(url)
		if err != nil {
			return nil
		}
		result.ParentURL = parent // ✅ 设置来源
		results = append(results, result)

		for _, link := range links {
			if sameDomain(link, targetURL) {
				_ = scanRecursive(link, url, currentDepth+1)
			}
		}
		return nil
	}

	_ = scanRecursive(targetURL, "", 0)
	return results, nil
}

func sameDomain(link1, link2 string) bool {
	u1, err1 := url.Parse(link1)
	u2, err2 := url.Parse(link2)
	if err1 != nil || err2 != nil {
		return false
	}
	return u1.Host == u2.Host
}

func (s *ScanSession) scanSingle(targetURL string) (ScanResult, []string, error) {
	page, err := s.context.NewPage()
	if err != nil {
		return ScanResult{}, nil, fmt.Errorf("创建页面失败: %w", err)
	}
	defer page.Close()

	var htmlList, jsList, cssList, imgList, otherList []string
	var apiList []APIEntry
	var aLinks []string
	seen := map[string]bool{}
	responseMap := make(map[string]playwright.Response)

	addURL := func(t string, url string) {
		if seen[url] {
			return
		}
		seen[url] = true
		switch t {
		case "html":
			htmlList = append(htmlList, url)
		case "js":
			jsList = append(jsList, url)
		case "css":
			cssList = append(cssList, url)
		case "image":
			imgList = append(imgList, url)
		case "other":
			otherList = append(otherList, url)
		}
	}

	page.On("request", func(req playwright.Request) {
		url := req.URL()
		method := req.Method()
		typ := detectResourceType(url, req.ResourceType(), method)

		if ClassifyAPI(req) {
			entry := APIEntry{
				URL:        url,
				Method:     method,
				PostData:   "",
				FromScript: true,
			}
			if method == "POST" {
				if body, err := req.PostData(); err == nil {
					entry.PostData = body
				}
			}
			apiList = append(apiList, entry)
		} else {
			addURL(typ, url)
		}
	})

	page.On("response", func(res playwright.Response) {
		url := res.URL()
		responseMap[url] = res
	})

	resp, err := page.Goto(targetURL)
	if err != nil {
		return ScanResult{}, nil, fmt.Errorf("页面加载失败: %w", err)
	}
	page.WaitForLoadState()
	page.WaitForTimeout(2000)

	if domUrls, err := ExtractDOMResources(page); err == nil {
		for _, u := range domUrls {
			typ := detectResourceType(u, "", "")
			addURL(typ, u)
		}
	}

	if hrefs, err := page.Evaluate(`() => {
		return [...document.querySelectorAll("a[href]")]
			.map(a => a.href).filter(h => h.startsWith("http"))
	}`); err == nil {
		if arr, ok := hrefs.([]interface{}); ok {
			for _, raw := range arr {
				if str, ok := raw.(string); ok {
					aLinks = append(aLinks, str)
				}
			}
		}
	}

	for _, frame := range page.Frames() {
		if frame == nil || frame.URL() == targetURL {
			continue
		}
		if urls, err := ExtractDOMResources(frame); err == nil {
			for _, u := range urls {
				typ := detectResourceType(u, "", "")
				addURL(typ, u)
			}
		}
	}

	for i, api := range apiList {
		if res, ok := responseMap[api.URL]; ok {
			apiList[i].Status = res.Status()
			apiList[i].Mime = res.Headers()["content-type"]
			if strings.Contains(apiList[i].Mime, "application/json") {
				if body, err := res.Text(); err == nil {
					apiList[i].Response = body
				}
			}
		}
	}

	title, _ := page.Title()
	if strings.TrimSpace(title) == "" {
		title = "无标题"
	}

	status := 0
	if resp != nil {
		status = resp.Status()
	}

	result := ScanResult{
		URL:       targetURL,
		Timestamp: time.Now().Format(time.RFC3339),
		Title:     title,
		Status:    status,
		HTMLList:  htmlList,
		JSList:    jsList,
		CSSList:   cssList,
		ImageList: imgList,
		APIList:   apiList,
		OtherList: otherList,
	}
	result.Counts.HTML = len(htmlList)
	result.Counts.JS = len(jsList)
	result.Counts.CSS = len(cssList)
	result.Counts.Image = len(imgList)
	result.Counts.API = len(apiList)
	result.Counts.Other = len(otherList)

	return result, aLinks, nil
}

func detectResourceType(url, resourceType, method string) string {
	u := strings.ToLower(url)

	if resourceType == "document" {
		return "html"
	}
	if resourceType == "script" || strings.HasSuffix(u, ".js") {
		return "js"
	}
	if strings.HasSuffix(u, ".css") || strings.Contains(u, "/css/") {
		return "css"
	}
	if strings.HasSuffix(u, ".png") || strings.HasSuffix(u, ".jpg") || strings.Contains(u, "/img/") {
		return "image"
	}
	if resourceType == "xhr" || resourceType == "fetch" || method == "POST" {
		return "api"
	}
	return "other"
}
