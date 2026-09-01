package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

func benchmark(url string, concurrency int, requests int) {
	start := time.Now()
	var wg sync.WaitGroup
	client := &http.Client{}
	success := 0
	mu := sync.Mutex{}
	sem := make(chan struct{}, concurrency)

	for i := 0; i < requests; i++ {
		wg.Add(1)
		sem <- struct{}{}

		go func(i int) {
			defer wg.Done()
			resp, err := client.Get(url)
			if err != nil {
				fmt.Println("请求错误：", err)
			} else if resp.StatusCode == 200 {
				mu.Lock()
				success++
				if success%50 == 0 {
					fmt.Printf("✅ 已完成 %d/%d 请求\n", success, requests)
				}
				mu.Unlock()
			}
			if resp != nil {
				resp.Body.Close()
			}
			<-sem
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)
	fmt.Printf("\n测试完成！\nURL: %s\n总请求数: %d\n成功数: %d\n平均响应时间: %.4fs\n总耗时: %.2fs\n\n",
		url, requests, success, elapsed.Seconds()/float64(requests), elapsed.Seconds())
}

func main() {
	fmt.Println("开始测试：未使用Redis缓存接口")
	benchmark("http://localhost:3000/api/articles?cache=off", 50, 1000)

	fmt.Println("开始测试：使用Redis缓存接口")
	benchmark("http://localhost:3000/api/articles?cache=on", 50, 1000)
}
