package main

import (
	"bufio"
	"fmt"
	"github.com/KagamigawaMeguri/mag/gohttp"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// 检查文件名是否有通配符
func fileNameIsGlob(pattern string) bool {
	_, err := regexp.Compile(pattern)
	return err == nil
}

// 判断是否为文件
func isFile(path string) bool {
	f, err := os.Stat(path)
	return err == nil && f.Mode().IsRegular()
}

// 判断是否为文件夹
func isDir(path string) bool {
	f, err := os.Stat(path)
	return err == nil && f.Mode().IsDir()
}

// 逐行读取文件
func readLines(filename string) ([]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return []string{}, err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

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

// 逗号分割字符串转mapset[int]
func stringToSliceInt(s string) ([]int, error) {
	if s == "" {
		return nil, nil
	}
	var set []int
	for _, v := range strings.Split(s, ",") {
		vTrim := strings.TrimSpace(v)
		if i, err := strconv.Atoi(vTrim); err == nil {
			set = append(set, i)
		} else {
			return set, err
		}
	}

	return set, nil
}

// 探测scheme
func ProbeScheme(host string, client *gohttp.HTTPClient) (string, error) {
	if !strings.HasPrefix(host, "http") {
		re := regexp.MustCompile(`^[^/]+:(\d+)`)
		match := re.FindStringSubmatch(host)
		if len(match) < 2 {
			// 无端口，默认80端口
			host = "http://" + host
		} else {
			port, err2 := strconv.Atoi(match[1])
			if err2 != nil {
				return "", fmt.Errorf("failed to parse port: %s", err2)
			} else if port == 443 {
				// 443端口，默认https
				host = "https://" + host
			} else {
				// 存在其他端口，默认为http
				host = "http://" + host
			}
		}
		_, err := client.SimpleRequest(host, "HEAD")
		log.Infof("host: %s", host)
		if err != nil {
			if strings.Contains(err.Error(), "EOF") {
				//判断为EOF，重试
				resp2, err2 := client.SimpleRequest(host, "GET")
				if err2 == nil && resp2.StatusCode == http.StatusOK {
					return host, nil
				}
			}
			if strings.Contains(err.Error(), "Client.Timeout") {
				//超时
				return host, nil
			}
			if strings.Contains(err.Error(), errHttps) {
				//说明为http
				host = strings.Replace(host, "https", "http", 1)
				return host, nil
			} else if err != nil && strings.Contains(err.Error(), "An existing connection was forcibly closed by the remote host") {
				return host, nil
			}
			log.Error(err)
		}
	}
	return host, nil
}

// Deduplicate 去重
func Deduplicate(arr []string) []string {
	set := make(map[string]struct{}, len(arr))
	j := 0
	for _, v := range arr {
		_, ok := set[v]
		if ok {
			continue
		}
		set[v] = struct{}{}
		arr[j] = v
		j++
	}

	return arr[:j]
}
