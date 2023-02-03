package gohttp

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
type Response struct {
	client     *HTTPClient
	Request    Request
	Status     string
	StatusCode int
	headers    []string
	Body       []byte
	Err        error
	elapsed    time.Duration
}

// 返回请求包和返回包的字符串
func (r *Response) String() string {
	//更换为strings.Builder以求提高性能
	//经过benchmark测试，以下方式资源占用最少且性能最高
	//唯一缺点写起来不太优雅
	var b strings.Builder
	b.Grow(512 + len(r.Body)) //预先开辟空间，可应对大部分情况

	b.WriteString(r.Request.URL() + "\n\n")
	b.WriteString(r.client.method + " " + r.Request.path + " HTTP/1.1\n")
	for k, v := range r.client.headers {
		b.WriteString(k + ": " + v + "\n")
	}
	b.WriteString("\n\n")
	b.WriteString("HTTP/1.1 " + r.Status + "\n")
	for _, h := range r.headers {
		b.WriteString(h + "\n")
	}
	b.WriteString("\n")
	b.Write(r.Body)

	return b.String()
}

func (r *Response) StringNoHeaders() string {
	var b strings.Builder
	b.Write(r.Body)
	return b.String()
}

// Save 将请求包和返回包写入到输出目录
func (r Response) Save(pathPrefix string) (string, error) {

	content := r.String()
	checksum := MD5(content)
	basename := strings.ReplaceAll(r.Request.path, "/", "_")
	parts := []string{pathPrefix}
	dir := strings.Replace(r.Request.host, ":", "_", 1)
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

func (r Response) UniqueId() string {
	return MD5(r.client.method + r.Request.host + string(r.Body))
}
func (r Response) BodyId() string {
	return MD5(r.client.method + r.Request.host + strconv.Itoa(len(r.Body)))
}

func MD5(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}
