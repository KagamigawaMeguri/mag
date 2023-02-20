package gohttp

import (
	"crypto/tls"
	"fmt"
	"github.com/KagamigawaMeguri/mag/opt"
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

// NewHTTPClient 对参数进行一些包装，构造成HTTPClient
func NewHTTPClient(options *opt.Options) (*HTTPClient, error) {
	var hc HTTPClient
	var err error
	var proxyURLFunc func(*http.Request) (*url.URL, error)
	if options.Proxy != "" {
		proxyURL, err := url.Parse(options.Proxy)
		if err != nil {
			return nil, fmt.Errorf("proxy URL is invalid (%w)", err)
		}
		proxyURLFunc = http.ProxyURL(proxyURL)
	} else {
		proxyURLFunc = nil
	}

	var redirectFunc func(req *http.Request, via []*http.Request) error
	if !options.FollowRedirects {
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
		Timeout:       time.Second * time.Duration(options.Timeout),
		CheckRedirect: redirectFunc,
		Transport: &http.Transport{
			DisableKeepAlives:   true,
			Proxy:               proxyURLFunc,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 0,
			TLSClientConfig:     &tlsConfig,
		}}
	hc.headers, err = options.Headers.TransformMap()
	if err != nil {
		return nil, err
	}
	if options.Method == "" {
		// 默认get
		hc.method = http.MethodGet
	} else {
		hc.method = options.Method
	}

	if options.Body == "" {
		hc.body = nil
	} else {
		hc.body = strings.NewReader(options.Body)
	}
	return &hc, nil
}

func (hc *HTTPClient) Request(r Request) (Response, error) {
	req, err := http.NewRequest(hc.method, r.URL(), hc.body)
	if err != nil {
		return Response{}, err
	}
	req.Host = r.Host
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
		return Response{}, err
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
	return Response{
		client:     hc,
		Request:    r,
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		headers:    hs,
		Body:       body,
	}, err
}
