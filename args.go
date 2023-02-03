package main

import (
	"fmt"
	"github.com/KagamigawaMeguri/mag/opt"
	"github.com/projectdiscovery/goflags"
	"regexp"
)

// 参数处理
func processArgs() (*opt.Options, error) {
	//接收参数
	options := &opt.Options{}
	flagSet := goflags.NewFlagSet()
	flagSet.SetDescription("兼顾效率与负载的多任务目录扫描器")

	flagSet.CreateGroup("input", "Input",
		flagSet.StringVarP(&options.Hosts, "list", "l", "./host.txt", "目标主机文件"),
		flagSet.StringVarP(&options.Paths, "path", "w", "./path.txt", "路径字典文件"),
	)

	flagSet.CreateGroup("output", "Output",
		flagSet.StringVarP(&options.Output, "output", "o", "./out", "输出路径"),
		flagSet.BoolVarP(&options.DisableOutput, "disableoutput", "do", false, "禁用输出"),
	)

	flagSet.CreateGroup("configs", "Configurations",
		flagSet.StringVarP(&options.Method, "method", "x", "", "自定义请求方法"),
		flagSet.StringVar(&options.Body, "body", "", "自定义请求包"),
		flagSet.StringVarP(&options.Proxy, "gohttp-proxy", "proxy", "", "设置代理 (eg http://127.0.0.1:8080)"),
		flagSet.VarP(&options.Headers, "header", "H", "自定义请求头"),
		flagSet.DurationVarP(&options.Delay, "delay", "d", 5*10000000, "扫描时相同host间最小延迟 (eg: 200ms, 1s)"),
		flagSet.IntVar(&options.Timeout, "timeout", 10, "请求超时时间"),
		flagSet.BoolVarP(&options.FollowRedirects, "follow", "f", true, "是否允许重定向"),
		flagSet.BoolVar(&options.Slow, "slow", false, "服务器极度友好模式"),
		flagSet.IntVarP(&options.Threads, "thread", "t", 25, "最大线程数"),
		flagSet.BoolVar(&options.RandomAgent, "random-agent", true, "是否启动随机UA-待开发"),
	)

	var (
		matchStatusCode string
		matchLength     string
		matchString     string
		matchRegex      string
	)
	flagSet.CreateGroup("matchers", "Matchers",
		flagSet.StringVarP(&matchStatusCode, "match-code", "mc", "", "匹配指定状态码 (eg: -mc 200,302)"),
		flagSet.StringVarP(&matchLength, "match-length", "ml", "", "匹配指定长度 (eg: -ml 100,102)"),
		flagSet.StringVarP(&matchString, "match-string", "ms", "", "匹配指定字符串 (eg: -ms admin)"),
		flagSet.StringVarP(&matchRegex, "match-regex", "mr", "", "匹配指定正则 (eg: -mr admin)"),
	)

	var (
		filterStatusCode string
		filterLength     string
		filterString     string
		filterRegex      string
	)
	flagSet.CreateGroup("filters", "Filters",
		flagSet.StringVarP(&filterStatusCode, "filter-code", "fc", "", "过滤指定状态码 (eg: -fc 403,401)"),
		flagSet.StringVarP(&filterLength, "filter-length", "fl", "", "过滤指定长度 (eg: -ml 100,102)"),
		flagSet.StringVarP(&filterString, "filter-string", "fs", "", "过滤指定长度 (eg: -fs admin)"),
		flagSet.StringVarP(&filterRegex, "filter-regex", "fr", "", "过滤指定正则 (eg: -fe admin)"),
	)

	flagSet.CreateGroup("debug", "Debug",
		flagSet.BoolVarP(&options.Verbose, "verbose", "v", false, "verbose mode"),
	)

	_ = flagSet.Parse()

	//参数校验与二次处理
	if options.Paths != "" && !fileNameIsGlob(options.Paths) && !isFile(options.Paths) {
		return options, fmt.Errorf("file '%s' does not exist", options.Paths)
	}

	if options.Hosts != "" && !fileNameIsGlob(options.Hosts) && !isFile(options.Hosts) {
		return options, fmt.Errorf("file '%s' does not exist", options.Hosts)
	}

	if options.Output != "" && !fileNameIsGlob(options.Output) && !isDir(options.Output) {
		return options, fmt.Errorf("'%s' is not a folder", options.Output)
	}

	var err error
	if options.MatchStatusCode, err = stringToMapsetInt(matchStatusCode); err != nil {
		return options, fmt.Errorf("invalid value for match status code option: %s", err)
	}

	if options.MatchLength, err = stringToMapsetInt(matchLength); err != nil {
		return options, fmt.Errorf("invalid value for match content length option: %s", err)
	}

	if matchRegex != "" {
		if options.MatchRegex, err = regexp.Compile(matchRegex); err != nil {
			return options, fmt.Errorf("invalid value for match regex option: %s", err)
		}
	}

	if options.FilterStatusCode, err = stringToMapsetInt(filterStatusCode); err != nil {
		return options, fmt.Errorf("invalid value for filter status code option: %s", err)
	}

	if options.FilterLength, err = stringToMapsetInt(filterLength); err != nil {
		return options, fmt.Errorf("invalid value for filter content length option: %s", err)
	}

	if filterRegex != "" {
		if options.FilterRegex, err = regexp.Compile(filterRegex); err != nil {
			return options, fmt.Errorf("invalid value for filter regex option: %s", err)
		}
	}

	return options, nil
}
