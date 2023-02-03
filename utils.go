package main

import (
	"bufio"
	mapset "github.com/deckarep/golang-set"
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

// 逗号分割字符串转mapset[int]
func stringToMapsetInt(s string) (mapset.Set, error) {
	set := mapset.NewSet[int]()
	if s == "" {
		return set, nil
	}
	for _, v := range strings.Split(s, ",") {
		vTrim := strings.TrimSpace(v)
		if i, err := strconv.Atoi(vTrim); err == nil {
			set.Add(i)
		} else {
			return set, err
		}
	}

	return set, nil
}
