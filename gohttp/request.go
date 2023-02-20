package gohttp

import (
	"fmt"
	"net/url"
)

// 请求包结构体
type Request struct {
	Scheme string
	Path   string
	Host   string
}

func NewRequest(host string, path string) (Request, error) {
	var r Request
	u, err := url.Parse(host)
	if err != nil {
		return r, fmt.Errorf("failed to parse Host: %s", err)
	}
	//路径处理
	prefixedPath, _ := url.JoinPath(u.Path, path)
	host, _ = url.JoinPath(host, prefixedPath)

	r = Request{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   prefixedPath,
	}
	return r, nil
}

// Hostname 返回hostname，http://foo.com:8080/ -> foo.com
func (r Request) Hostname() string {
	u, err := url.Parse(r.URL())

	if err != nil {
		return "unknown"
	}
	return u.Hostname()
}

// URL 返回完整url
func (r Request) URL() string {
	target, _ := url.JoinPath(r.Host, r.Path)
	return r.Scheme + "://" + target
}

// DeepCopy 深拷贝
func DeepCopy(src Request) Request {
	dst := Request{
		Scheme: src.Scheme,
		Path:   src.Path,
		Host:   src.Host,
	}
	return dst
}
