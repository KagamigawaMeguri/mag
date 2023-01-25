package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
)

type multiStringArgs []string

func (h *multiStringArgs) Set(val string) error {
	for _, v := range strings.Split(val, ",") {
		*h = append(*h, v)
	}
	return nil
}

func (h *multiStringArgs) String() string {
	return "string"
}

type multiIntArgs []int

func (s *multiIntArgs) Set(val string) error {
	for _, v := range strings.Split(val, ",") {
		i, _ := strconv.Atoi(v)
		*s = append(*s, i)
	}
	return nil
}

func (s *multiIntArgs) String() string {
	return "string"
}

func (s *multiIntArgs) Includes(search int) bool {
	for _, status := range *s {
		if status == search {
			return true
		}
	}
	return false
}

type config struct {
	method         string
	body           string
	threads        int
	delay          int
	headers        multiStringArgs
	followLocation bool
	matchString    string
	matchRegex     string
	matchCode      multiIntArgs
	filterLength   multiIntArgs
	filterString   string
	filterRegex    string
	timeout        int
	verbose        bool
	paths          string
	hosts          string
	output         string
	noHeaders      bool
	proxy          string
	slow           bool
}

func processArgs() config {

	method := "GET"
	flag.StringVar(&method, "method", "GET", "")
	flag.StringVar(&method, "x", "GET", "")

	var headers multiStringArgs
	flag.Var(&headers, "header", "")
	flag.Var(&headers, "H", "")

	body := ""
	flag.StringVar(&body, "body", "", "")
	flag.StringVar(&body, "b", "", "")

	threads := 20
	flag.IntVar(&threads, "threads", 20, "")
	flag.IntVar(&threads, "t", 20, "")

	delay := 5000
	flag.IntVar(&delay, "delay", 5000, "")
	flag.IntVar(&delay, "d", 5000, "")

	followRedirects := false
	flag.BoolVar(&followRedirects, "follow-redirects", false, "")
	flag.BoolVar(&followRedirects, "fr", false, "")

	matchRegex := ""
	flag.StringVar(&matchRegex, "match-regex", "", "")
	flag.StringVar(&matchRegex, "mr", "", "")

	matchString := ""
	flag.StringVar(&matchString, "match-string", "", "")
	flag.StringVar(&matchString, "ms", "", "")

	var matchCode multiIntArgs
	flag.Var(&matchCode, "match-code", "")
	flag.Var(&matchCode, "mc", "")

	filterString := ""
	flag.StringVar(&filterString, "filter-string", "", "")
	flag.StringVar(&filterString, "fs", "", "")

	filterRegex := ""
	flag.StringVar(&filterRegex, "filter-regex", "", "")
	flag.StringVar(&filterRegex, "fe", "", "")

	var filterLength multiIntArgs
	flag.Var(&filterLength, "filter-length", "")
	flag.Var(&filterLength, "fl", "")

	timeout := 10000
	flag.IntVar(&timeout, "timeout", 10000, "")

	var slow bool
	flag.BoolVar(&slow, "slow", false, "")

	noHeaders := false
	flag.BoolVar(&noHeaders, "no-headers", false, "")

	verbose := false
	flag.BoolVar(&verbose, "verbose", false, "")
	flag.BoolVar(&verbose, "v", false, "")

	var proxy string
	flag.StringVar(&proxy, "http-proxy", "", "")
	flag.StringVar(&proxy, "proxy", "", "")

	flag.Parse()

	// path路径
	paths := flag.Arg(0)
	if paths == "" {
		paths = defaultPathsFile
	}

	// host路径
	hosts := flag.Arg(1)
	if hosts == "" {
		hosts = defaultHostsFile
	}

	// 输出路径
	output := flag.Arg(2)
	if output == "" {
		output = defaultOutputDir
	}

	return config{
		method:         method,
		body:           body,
		threads:        threads,
		delay:          delay,
		filterString:   filterString,
		headers:        headers,
		followLocation: followRedirects,
		filterRegex:    filterRegex,
		matchRegex:     matchRegex,
		matchString:    matchString,
		matchCode:      matchCode,
		timeout:        timeout,
		verbose:        verbose,
		paths:          paths,
		hosts:          hosts,
		output:         output,
		noHeaders:      noHeaders,
		filterLength:   filterLength,
		proxy:          proxy,
	}
}

func init() {
	flag.Usage = func() {
		h := "基于path的服务器友好型多host目录扫描器\n"

		h += "\n用法:\n"
		h += "  mag [pathsFile] [hostsFile] [outputDir]\n"

		h += "\n请求:\n"
		h += fmt.Sprintf("%-36s\t%s\n", "  -X, -method <string>", "设置请求方法(默认GET)")
		h += fmt.Sprintf("%-36s\t%s\n", "  -H, -header <string>", "设置请求头")
		h += fmt.Sprintf("%-36s\t%s\n", "  -b, -body <string> ", "设置POST请求体")
		h += fmt.Sprintf("%-36s\t%s\n", "  -t, -threads <int>", "设置并发数(默认20)")
		h += fmt.Sprintf("%-36s\t%s\n", "  -d, -delay <int>", "设置相同host间的延迟(默认5000ms)")
		h += fmt.Sprintf("%-36s\t%s\n", "  -timeout <int>", "设置超时时间(默认10000ms)")
		h += fmt.Sprintf("%-36s\t%s\n", "  -proxy <string>", "设置代理")
		h += fmt.Sprintf("%-36s\t%s\n", "  -fr, -follow-redirects", "允许重定向")
		h += fmt.Sprintf("%-36s\t%s\n", "  -no-headers", "不设置请求头")
		h += fmt.Sprintf("%-36s\t%s\n", "  -slow", "服务器极度友好模式")

		h += "\n匹配:\n"
		h += fmt.Sprintf("%-36s\t%s\n", "  -ms, -match-string <string>", "检测到指定字符串则保存")
		h += fmt.Sprintf("%-36s\t%s\n", "  -mr, -match-regex <string>", "检测到指定regex则保存")
		h += fmt.Sprintf("%-36s\t%s\n", "  -mc, -match-code <int>", "检测到指定状态码则保存：-match-code 200,301")

		h += "\n过滤:\n"
		h += fmt.Sprintf("%-36s\t%s\n", "  -fs, -filter-string <string>", "检测到指定字符串则跳过")
		h += fmt.Sprintf("%-36s\t%s\n", "  -fe, -filter-regex <string>", "检测到指定regex则跳过")
		h += fmt.Sprintf("%-36s\t%s\n", "  -fl, -filter-length <int>", "检测到指定长度则跳过：-match-code 200,301")

		h += "\nDEBUG:\n"
		h += fmt.Sprintf("%-36s\t%s\n", "  -v,  -verbose", "Verbose mode")

		h += "\n默认路径:\n"
		h += "  pathsFile: ./paths\n"
		h += "  hostsFile: ./hosts\n"
		h += "  outputDir: ./out\n"

		fmt.Println(h)
	}
}
