package main

import (
	"fmt"
	"github.com/KagamigawaMeguri/mag/gohttp"
	"github.com/KagamigawaMeguri/mag/opt"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// 自定义错误
const (
	errTls = "tls: server selected unsupported protocol version 301"
)

var (
	simpleFilter = NewSimpleFilter(uint8(25))
	logger, _    = initLogger()
	log          = logger.Sugar()
)

func initiate(c *opt.Options) ([]string, []string, *os.File) {
	// 读path文件
	paths, err := readLines(c.Paths)
	if err != nil {
		log.Fatalf("failed to open paths file: %s", err)
		os.Exit(1)
	}

	// 读host文件
	hosts, err := readLines(c.Hosts)
	if err != nil {
		log.Fatalf("failed to open hosts file: %s", err)
		os.Exit(1)
	}

	// 创建输出目录
	err = os.MkdirAll(c.Output, 0750)
	if err != nil {
		log.Fatalf("failed to create output directory: %s", err)
		os.Exit(1)
	}

	// 创建index文件
	indexFile := filepath.Join(c.Output, "index")
	index, err := os.OpenFile(indexFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatalf("failed to open index file for writing: %s", err)
		os.Exit(1)
	}
	return hosts, paths, index
}

func main() {
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {
		}
	}(logger)
	var err error
	// 获取参数配置
	options, err := processArgs()
	if err != nil {
		log.Fatal(err)
	}
	hosts, paths, index := initiate(options)

	// 打印任务情况
	fmt.Printf("[+] Urllist: %s\n"+
		"[+] Method: %s\n"+
		"[+] Threads: %d\n"+
		"[+] Pathlist: %s\n"+
		"[+] Timeout: %d\n", options.Hosts, options.Method, options.Threads, options.Paths, options.Delay)

	// 设置限速器
	rl := newRateLimiter(time.Duration(options.Delay*1000000), options.Slow)

	requestsChan := make(chan gohttp.Request)
	responsesChan := make(chan gohttp.Response)
	client, err := gohttp.NewHTTPClient(options)

	// 请求处理
	var wg sync.WaitGroup
	for i := 0; i < options.Threads; i++ {
		wg.Add(1)
		//不使用闭包，以求减少资源
		go func(items chan gohttp.Request) {
			for r := range items {
				rl.Block(r.Hostname()) //传入限速器判断是否限速
				ret, err := client.Request(r)
				if err != nil {
					log.Warn(err)
				}
				responsesChan <- ret //发送请求，返回包写入responses channel
			}
			wg.Done()
		}(requestsChan)
	}

	// 返回包处理
	var owg sync.WaitGroup
	owg.Add(1)
	go func(items chan gohttp.Response) {
		for resp := range items {
			if resp.Err != nil {
				if !strings.Contains(err.Error(), errTls) {
					log.Infof("%s", err)
				}
				continue
			}

			if !options.MatchStatusCode.Contains(resp.StatusCode) {
				continue
			}
			if !options.MatchLength.Contains(len(resp.Body)) {
				continue
			}

			if !(options.MatchRegex.String() == "") && !options.MatchRegex.Match(resp.Body) {
				continue
			}

			if options.FilterStatusCode.Contains(resp.StatusCode) {
				continue
			}
			if options.FilterLength.Contains(len(resp.Body)) {
				continue
			}

			if !(options.FilterRegex.String() == "") && options.FilterRegex.Match(resp.Body) {
				continue
			}

			//过滤重复回显
			if simpleFilter.DoFilter(resp) {
				continue
			}
			path, err := resp.Save(options.Output)
			if err != nil {
				log.Infof("failed to save file: %s", err)
			}

			line := fmt.Sprintf("%s %s [%s] [%d]", path, resp.Request.URL(), resp.Status, len(resp.Body))
			fmt.Fprintf(index, line)
			if options.Verbose {
				log.Infof("%s", line)
			}
		}
		owg.Done()
	}(responsesChan)
	if err != nil {
		log.Fatal(err)
	}
	// 基于path去添加请求
	for _, path := range paths {
		for _, host := range hosts {
			r, err := gohttp.NewRequest(host, path)
			if err != nil {
				log.Warn(err)
				continue
			}
			requestsChan <- r
		}
	}

	close(requestsChan)
	wg.Wait()

	close(responsesChan)
	owg.Wait()

}

func initLogger() (*zap.Logger, error) {
	cfg := zap.NewDevelopmentConfig()
	cfg.DisableStacktrace = true
	cfg.DisableCaller = true
	return cfg.Build()
}
