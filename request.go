package main

import (
	"net/url"
)

// 请求包结构体
type request struct {
	scheme string
	path   string
	host   string
}

// Hostname 返回hostname，http://foo.com:8080 -> foo.com
func (r request) Hostname() string {
	u, err := url.Parse(r.URL())

	if err != nil {
		return "unknown"
	}
	return u.Hostname()
}

// URL 返回完整url
func (r request) URL() string {
	target, _ := url.JoinPath(r.host, r.path)
	return r.scheme + "://" + target
}
