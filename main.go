package main

import (
	"fmt"
	"github.com/KagamigawaMeguri/mag/gohttp"
	"github.com/KagamigawaMeguri/mag/lib"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// 自定义错误
const (
	errTls   = "tls: server selected unsupported protocol version 301"
	errHttps = "http: server gave HTTP response to HTTPS client"
)

var (
	simpleFilter = NewSimpleFilter(uint8(25))
	logger, _    = initLogger()
	log          = logger.Sugar()
)

func initiate(c *lib.Options) ([]string, []string, *os.File) {
	// 读path文件
	paths, err := readLines(c.Paths)
	paths = Deduplicate(paths)
	if err != nil {
		log.Fatalf("failed to open paths file: %s", err)
	}
	// 读host文件
	hosts, err := readLines(c.Hosts)
	hosts = Deduplicate(hosts)
	if err != nil {
		log.Fatalf("failed to open hosts file: %s", err)
	}
	// 探测协议与二次清洗
	httpOpt := &lib.HttpOptions{
		Method:          "HEAD",
		Timeout:         c.Timeout,
		FollowRedirects: true,
		Proxy:           c.Proxy,
	}
	client, _ := gohttp.NewHTTPClient(httpOpt)
	poolSize := c.Threads
	pool := make(chan struct{}, poolSize)
	hostChan := make(chan string)
	var newHosts []string
	go func() {
		for i := range hostChan {
			newHosts = append(newHosts, i)
		}
	}()
	var wg sync.WaitGroup
	wg.Add(len(hosts))
	for _, url := range hosts {
		pool <- struct{}{}
		go func(u string) {
			defer wg.Done()
			host, err := ProbeScheme(u, client)
			if err != nil {
				log.Error(err)
			}
			hostChan <- host
			<-pool
		}(url)
	}
	wg.Wait()
	newHosts = Deduplicate(newHosts)
	log.Infof("fix hosts successfully...")
	// 创建输出目录
	err = os.MkdirAll(c.Output, 0750)
	if err != nil {
		log.Fatalf("failed to create output directory: %s", err)
	}

	// 创建index文件
	indexFile := filepath.Join(c.Output, "index")
	index, err := os.OpenFile(indexFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatalf("failed to open index file for writing: %s", err)
	}
	return newHosts, paths, index
}

func main() {
	defer func(logger *zap.Logger) {
		_ = logger.Sync()
	}(logger)
	var err error
	// 获取参数配置
	options, err := processArgs()
	if err != nil {
		log.Fatal(err)
	}
	hosts, paths, index := initiate(options)

	// 打印任务情况
	fmt.Printf("[+] Hostlist: %s\n"+
		"[+] Method: %s\n"+
		"[+] Threads: %d\n"+
		"[+] Pathlist: %s\n"+
		"[+] Timeout: %s\n", options.Hosts, options.Method, options.Threads, options.Paths, options.Delay)

	// 设置限速器
	rl := newRateLimiter(options.Delay, options.Slow)

	requestsChan := make(chan gohttp.Request)
	responsesChan := make(chan gohttp.Response)
	client, err := gohttp.NewHTTPClient(gohttp.ParseHttpOptions(options))

	// 请求处理
	var wg sync.WaitGroup
	for i := 0; i < options.Threads; i++ {
		wg.Add(1)
		//不使用闭包，以求减少资源
		go func(items chan gohttp.Request) {
			for r := range items {
				rl.Block(r.Hostname()) //传入限速器判断是否限速
				if err != nil {
					log.Error(err)
					continue
				}
				ret, err := client.Request(r)
				if err != nil {
					log.Error(err)
					continue
				}
				if options.Verbose {
					log.Debug(r.URL())
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
			//if options.Verbose {
			//	log.Debugf("%s [%d] [%d]", resp.Request.URL(), resp.StatusCode, len(resp.Body))
			//}
			//默认保存200
			if (options.MatchStatusCode != nil && !(slices.Contains(options.MatchStatusCode, resp.StatusCode))) || !slices.Contains(options.MatchStatusCode, http.StatusOK) {
				continue
			}

			if options.MatchLength != nil && !(slices.Contains(options.MatchLength, len(resp.Body))) {
				continue
			}

			if options.FilterRegex != nil && !options.MatchRegex.Match(resp.Body) {
				continue
			}

			if options.FilterStatusCode != nil && slices.Contains(options.FilterStatusCode, resp.StatusCode) {
				continue
			}
			if options.FilterLength != nil && slices.Contains(options.FilterLength, len(resp.Body)) {
				continue
			}

			if options.FilterRegex != nil && options.FilterRegex.Match(resp.Body) {
				continue
			}

			//过滤重复回显
			if simpleFilter.DoFilter(&resp) {
				continue
			}
			path, err := resp.Save(options.Output)
			if err != nil {
				log.Infof("failed to save file: %s", err)
			}

			line := fmt.Sprintf("%s %s [%d] [%d]", path, resp.Request.URL(), resp.StatusCode, len(resp.Body))
			_, err = fmt.Fprintf(index, line)
			if err != nil {
				log.Fatalf("failed to write to index file: %s", err)
			}
			log.Infof("%s", line)
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
			//根据host增加备份文件路径, www.rar -> foo.com.www.rar
			if options.Backup {
				r2 := gohttp.DeepCopy(r)
				r2.Path = r2.Hostname() + "." + r2.Path
				if err != nil {
					log.Warn(err)
					continue
				}
				requestsChan <- r2
			}
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
	cfg.DisableCaller = false
	return cfg.Build()
}
