package main

import (
	"crypto/md5"
	"encoding/hex"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

// 返回包结构体
type response struct {
	client     *HTTPClient
	request    request
	status     string
	statusCode int
	headers    []string
	body       []byte
	err        error
	elapsed    time.Duration
}

// 返回请求包和返回包的字符串
func (r response) String() string {
	//更换为strings.Builder以求提高性能
	//经过benchmark测试，以下方式资源占用最少且性能最高
	//唯一缺点写起来不太优雅
	var b strings.Builder
	b.Grow(512 + len(r.body)) //预先开辟空间，可应对大部分情况

	b.WriteString(r.request.URL() + "\n\n")
	b.WriteString(r.client.method + " " + r.request.path + " HTTP/1.1\n")
	for k, v := range r.client.headers {
		b.WriteString(k + ": " + v + "\n")
	}
	b.WriteString("\n\n")
	b.WriteString("HTTP/1.1 " + r.status + "\n")
	for _, h := range r.headers {
		b.WriteString(h + "\n")
	}
	b.WriteString("\n")
	b.Write(r.body)

	return b.String()
}

func (r response) StringNoHeaders() string {
	var b strings.Builder
	b.Write(r.body)
	return b.String()
}

// save 将请求包和返回包写入到输出目录
func (r response) save(pathPrefix string, noHeaders bool) (string, error) {

	content := r.String()
	if noHeaders {
		content = r.StringNoHeaders()
	}

	checksum := MD5(content)
	basename := strings.ReplaceAll(r.request.path, "/", "_")
	parts := []string{pathPrefix}
	dir := strings.Replace(r.request.host, ":", "_", 1)
	parts = append(parts, dir)
	parts = append(parts, basename+"_"+checksum[:6])

	p := path.Join(parts...)

	if _, err := os.Stat(path.Dir(p)); os.IsNotExist(err) {
		err = os.MkdirAll(path.Dir(p), 0750)
		if err != nil {
			return p, err
		}
	}

	err := os.WriteFile(p, []byte(content), 0640)
	if err != nil {
		return p, err
	}

	return p, nil
}

func (r response) uniqueId() string {
	return MD5(r.client.method + r.request.host + string(r.body))
}
func (r response) bodyId() string {
	return MD5(r.client.method + r.request.host + strconv.Itoa(len(r.body)))
}

func MD5(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}
