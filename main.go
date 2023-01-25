package main

import (
	"bufio"
	"fmt"
	"go.uber.org/zap"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// 默认参数
const (
	defaultPathsFile = "./paths"
	defaultHostsFile = "./hosts"
	defaultOutputDir = "./out"
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

func initiate(c *config) ([]string, []string, *os.File) {
	// 读path文件
	paths, err := readLinesOrLiteral(c.paths, defaultPathsFile)
	if err != nil {
		log.Fatalf("failed to open paths file: %s", err)
		os.Exit(1)
	}

	// 读host文件
	hosts, err := readLinesOrLiteral(c.hosts, defaultHostsFile)
	if err != nil {
		log.Fatalf("failed to open hosts file: %s", err)
		os.Exit(1)
	}

	// 创建输出目录
	err = os.MkdirAll(c.output, 0750)
	if err != nil {
		log.Fatalf("failed to create output directory: %s", err)
		os.Exit(1)
	}

	// 创建index文件
	indexFile := filepath.Join(c.output, "index")
	index, err := os.OpenFile(indexFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatalf("failed to open index file for writing: %s", err)
		os.Exit(1)
	}
	return hosts, paths, index
}

func main() {
	defer logger.Sync()
	var err error
	// 获取参数配置
	c := processArgs()
	hosts, paths, index := initiate(&c)

	// 打印任务情况
	fmt.Printf("[+] Urllist: %s\n"+
		"[+] Method: %s\n"+
		"[+] Threads: %d\n"+
		"[+] Wordlist: %s\n"+
		"[+] Timeout: %d\n", c.hosts, c.method, c.threads, c.paths, c.delay)

	// 设置限速器
	rl := newRateLimiter(time.Duration(c.delay*1000000), c.slow)

	requestsChan := make(chan request)
	responsesChan := make(chan response)
	client, err := newHTTPClient(c)

	// 请求处理
	var wg sync.WaitGroup
	for i := 0; i < c.threads; i++ {
		wg.Add(1)
		//不使用闭包，以求减少资源
		go func(items chan request) {
			for r := range items {
				rl.Block(r.Hostname()) //传入限速器判断是否限速
				ret, err := client.request(r)
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
	go func(items chan response) {
		for res := range items {
			if len(c.matchCode) > 0 && !c.matchCode.Includes(res.statusCode) {
				continue
			}

			if len(c.filterLength) > 0 && c.matchCode.Includes(len(res.body)) {
				continue
			}

			if len(c.matchString) > 0 && !(strings.Contains(strings.Join(res.headers, ""), c.matchString) || (strings.Contains(string(res.body), c.matchString))) {
				continue
			}

			if len(c.filterString) > 0 && (strings.Contains(strings.Join(res.headers, ""), c.filterString) || (strings.Contains(string(res.body), c.filterString))) {
				continue
			}
			if len(c.filterRegex) > 0 {
				re := regexp.MustCompile(c.filterRegex)
				matched := re.MatchString(res.String())
				if matched {
					continue
				}
			}

			if len(c.matchRegex) > 0 {
				re := regexp.MustCompile(c.matchRegex)
				matched := re.MatchString(res.String())
				if !matched {
					continue
				}
			}

			if res.err != nil {
				if !strings.Contains(err.Error(), errTls) {
					log.Infof("%s", err)
				}
				continue
			}

			//过滤重复回显
			if simpleFilter.DoFilter(res) {
				continue
			}
			path, err := res.save(c.output, c.noHeaders)
			if err != nil {
				log.Infof("failed to save file: %s", err)
			}

			line := fmt.Sprintf("%s %s [%s] [%d]", path, res.request.URL(), res.status, len(res.body))
			fmt.Fprintf(index, line)
			if c.verbose {
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
			target, err := addScheme(host)
			if err != nil {
				log.Warn(err)
				continue
			}
			u, err := url.Parse(target)
			if err != nil {
				log.Warn("failed to parse host: %s", err)
				continue
			}
			//路径处理
			prefixedPath, _ := url.JoinPath(u.Path, path)
			target, _ = url.JoinPath(target, prefixedPath)

			requestsChan <- request{
				scheme: u.Scheme,
				host:   u.Host,
				path:   prefixedPath,
			}
		}
	}

	close(requestsChan)
	wg.Wait()

	close(responsesChan)
	owg.Wait()

}

// 逐行读取文件
func readLines(filename string) ([]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return []string{}, err
	}
	defer f.Close()

	lines := make([]string, 0)
	sc := bufio.NewScanner(f)
	var s string
	for sc.Scan() {
		s = sc.Text()
		//跳过空行
		if s == "\n" || s == "\r\n" {
			continue
		}
		lines = append(lines, s)
	}

	return lines, sc.Err()
}

// readLinesOrLiteral 读取文件
func readLinesOrLiteral(arg, argDefault string) ([]string, error) {
	if isFile(arg) {
		return readLines(arg)
	}

	if arg != argDefault {
		return []string{}, fmt.Errorf("file %s not found", arg)
	}
	return readLines(argDefault)

}

// 判断是否为文件
func isFile(path string) bool {
	f, err := os.Stat(path)
	return err == nil && f.Mode().IsRegular()
}

func addScheme(url string) (string, error) {
	// 无scheme处理
	if !strings.HasPrefix(url, "http") {
		re := regexp.MustCompile(`^[^/]+:(\d+)`)
		match := re.FindStringSubmatch(url)
		if len(match) < 2 {
			// 无端口，默认80端口
			url = "http://" + url
		} else {
			port, err2 := strconv.Atoi(match[1])
			if err2 != nil || (port != 80 && port != 443) {
				return "", fmt.Errorf("url scheme not specified")
			} else if port == 80 {
				url = "http://" + url
			} else {
				url = "https://" + url
			}
		}
	}
	return url, nil
}

func initLogger() (*zap.Logger, error) {
	cfg := zap.NewDevelopmentConfig()
	cfg.DisableStacktrace = true
	cfg.DisableCaller = true
	return cfg.Build()
}
