package main

import (
	"bufio"
	"fmt"
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
func readLines(filename string, mode ...string) ([]string, error) {
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
		if len(mode) > 0 && mode[0] == "host" {
			// 增加scheme
			if host, err := addScheme(s); err == nil {
				lines = append(lines, host)
			} else {
				return nil, err
			}
		} else {
			lines = append(lines, s)
		}

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

// 增加scheme
func addScheme(host string) (string, error) {
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
