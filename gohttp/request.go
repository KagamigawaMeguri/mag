package gohttp

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// 请求包结构体
type Request struct {
	scheme string
	path   string
	host   string
}

func NewRequest(host string, path string) (Request, error) {
	var r Request
	target := host
	if !strings.HasPrefix(target, "gohttp") {
		re := regexp.MustCompile(`^[^/]+:(\d+)`)
		match := re.FindStringSubmatch(target)
		if len(match) < 2 {
			// 无端口，默认80端口
			target = "http://" + target
		} else {
			port, err2 := strconv.Atoi(match[1])
			if err2 != nil || (port != 80 && port != 443) {
				return r, fmt.Errorf("target scheme not specified")
			} else if port == 80 {
				target = "http://" + target
			} else {
				target = "https://" + target
			}
		}
	}
	u, err := url.Parse(target)
	if err != nil {
		return r, fmt.Errorf("failed to parse host: %s", err)
	}
	//路径处理
	prefixedPath, _ := url.JoinPath(u.Path, path)
	target, _ = url.JoinPath(target, prefixedPath)

	r = Request{
		scheme: u.Scheme,
		host:   u.Host,
		path:   prefixedPath,
	}
	return r, nil
}

// Hostname 返回hostname，http://foo.com:8080 -> foo.com
func (r Request) Hostname() string {
	u, err := url.Parse(r.URL())

	if err != nil {
		return "unknown"
	}
	return u.Hostname()
}

// URL 返回完整url
func (r Request) URL() string {
	target, _ := url.JoinPath(r.host, r.path)
	return r.scheme + "://" + target
}
