package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"web_scanner/internal"
	"web_scanner/internal/server"
)

func main() {
	urlFlag := flag.String("url", "", "单个 URL")
	fileFlag := flag.String("file", "", "批量 URL 文件路径")
	outFlag := flag.String("out", "report/report.json", "输出 JSON 路径")
	threads := flag.Int("threads", 4, "并发线程数")
	depth := flag.Int("depth", 0, "最大递归深度（默认 0 只扫当前页）")
	flag.Parse()

	var urls []string
	if *urlFlag != "" {
		urls = append(urls, *urlFlag)
	}
	if *fileFlag != "" {
		f, err := os.Open(*fileFlag)
		if err != nil {
			fmt.Printf("❌ 无法读取文件: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				urls = append(urls, line)
			}
		}
	}

	if len(urls) == 0 {
		fmt.Println("请使用 -url 或 -file 提供至少一个地址")
		return
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var allResults []internal.ScanResult
	tasks := make(chan string)

	for i := 0; i < *threads; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for url := range tasks {
				fmt.Printf("线程 %d 扫描中: %s\n", id, url)

				session, err := internal.InitScanSession()
				if err != nil {
					fmt.Printf("❌ 启动失败: %v\n", err)
					continue
				}

				results, err := session.RunScan(url, *depth)
				if err != nil {
					fmt.Printf("❌ 扫描失败: %v\n", err)
					session.Close()
					continue
				}
				session.Close()

				mu.Lock()
				allResults = append(allResults, results...)
				mu.Unlock()

				fmt.Printf("✅ 完成: %s（共 %d 页）\n", url, len(results))
			}
		}(i)
	}

	for _, u := range urls {
		tasks <- u
	}
	close(tasks)
	wg.Wait()

	err := internal.SaveResultsToJSON(allResults, *outFlag)
	if err != nil {
		fmt.Printf("❌ 报告生成失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ JSON 报告已生成: %s\n", *outFlag)
	server.StartReportServer("report")
	server.OpenBrowser("http://localhost:8080")

	// 防止主线程退出
	select {}
}
