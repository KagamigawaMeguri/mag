package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// 最大获取100K的响应，适用于绝大部分场景
const defaultResponseLength = 10240

var (
	//默认头
	defaultHeaders = map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36 Edg/108.0.1462.76",
		"Range":      fmt.Sprintf("bytes=0-%d", defaultResponseLength),
	}
)

type HTTPClient struct {
	client    *http.Client
	body      io.Reader
	userAgent string
	headers   map[string]string
	method    string
}

// 对参数进行一些包装，构造成HTTPClient
func newHTTPClient(c config) (*HTTPClient, error) {
	var hc HTTPClient

	var proxyURLFunc func(*http.Request) (*url.URL, error)
	if c.proxy != "" {
		proxyURL, err := url.Parse(c.proxy)
		if err != nil {
			return nil, fmt.Errorf("proxy URL is invalid (%w)", err)
		}
		proxyURLFunc = http.ProxyURL(proxyURL)
	} else {
		proxyURLFunc = nil
	}

	var redirectFunc func(req *http.Request, via []*http.Request) error
	if !c.followLocation {
		redirectFunc = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	} else {
		redirectFunc = nil
	}
	tlsConfig := tls.Config{
		InsecureSkipVerify: true,
		// enable TLS1.0 and TLS1.1 support
		MinVersion: tls.VersionTLS10,
	}
	hc.client = &http.Client{
		Timeout:       time.Duration(c.timeout * 1000000),
		CheckRedirect: redirectFunc,
		Transport: &http.Transport{
			Proxy:               proxyURLFunc,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 0,
			TLSClientConfig:     &tlsConfig,
		}}
	if !c.noHeaders {
		for _, h := range c.headers {
			keyAndValue := strings.SplitN(h, ":", 2)
			if len(keyAndValue) != 2 {
				return nil, fmt.Errorf("invalid header format for header %q", h)
			}
			key := strings.TrimSpace(keyAndValue[0])
			value := strings.TrimSpace(keyAndValue[1])
			if len(key) == 0 {
				return nil, fmt.Errorf("invalid header format for header %q - name is empty", h)
			}
			hc.headers[key] = value
		}
	}
	if c.method == "" {
		// 默认get
		hc.method = http.MethodGet
	} else {
		hc.method = c.method
	}

	if c.body == "" {
		hc.body = nil
	} else {
		hc.body = strings.NewReader(c.body)
	}
	return &hc, nil
}

func (hc *HTTPClient) request(r request) (response, error) {
	req, err := http.NewRequest(hc.method, r.URL(), hc.body)
	if err != nil {
		return response{request: r}, err
	}
	req.Host = r.host
	//设置自定义header
	for k, v := range hc.headers {
		req.Header.Set(k, v)
	}
	// 补充默认header
	for k, v := range defaultHeaders {
		if req.Header.Get(k) == "" {
			req.Header.Set(k, v)
		}
	}
	resp, err := hc.client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return response{request: r}, err
	}
	// 提取响应头
	hs := make([]string, 0)
	for k, vs := range resp.Header {
		for _, v := range vs {
			hs = append(hs, k+v)
		}
	}
	body, _ := io.ReadAll(resp.Body)
	// 带Range头后一般webserver响应都是[206 PARTIAL CONTENT]，修正为[200 OK]
	if resp.StatusCode == 206 {
		resp.StatusCode = 200
		resp.Status = "200 OK"
	}
	return response{
		client:     hc,
		request:    r,
		status:     resp.Status,
		statusCode: resp.StatusCode,
		headers:    hs,
		body:       body,
	}, err
}
