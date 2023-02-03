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
		flagSet.StringVarP(&options.Hosts, "host", "h", "./host.txt", "input file containing list of hosts to process"),
		flagSet.StringVarP(&options.Paths, "path", "p", "./path.txt", "input file containing list of paths to process"),
	)

	flagSet.CreateGroup("output", "Output",
		flagSet.StringVarP(&options.Output, "output", "o", "./out", "dir to write output results"),
		flagSet.BoolVarP(&options.DisableOutput, "disableoutput", "do", false, "disable output"),
	)

	flagSet.CreateGroup("configs", "Configurations",
		flagSet.StringVarP(&options.Method, "method", "x", "", "gohttp methods to probe"),
		flagSet.StringVar(&options.Body, "body", "", "post body to include in gohttp gohttp"),
		flagSet.StringVarP(&options.Proxy, "gohttp-proxy", "proxy", "", "gohttp proxy to use (eg http://127.0.0.1:8080)"),
		flagSet.VarP(&options.Headers, "header", "H", "custom gohttp headers to send with gohttp"),
		flagSet.DurationVarP(&options.Delay, "delay", "d", 5, "duration between each gohttp gohttp with same host (eg: 200ms, 1s)"),
		flagSet.IntVar(&options.Timeout, "timeout", 10, "timeout in seconds"),
		flagSet.BoolVarP(&options.FollowRedirects, "follow-redirects", "fr", false, "follow gohttp redirects"),
		flagSet.BoolVar(&options.Slow, "slow", false, "Server extremely friendly mode"),
		flagSet.IntVarP(&options.Threads, "thread", "t", 25, "number of threads to use"),
		flagSet.BoolVar(&options.RandomAgent, "random-agent", true, "enable Random User-Agent to use"),
	)

	var (
		matchStatusCode string
		matchLength     string
		matchString     string
		matchRegex      string
	)
	flagSet.CreateGroup("matchers", "Matchers",
		flagSet.StringVarP(&matchStatusCode, "match-code", "mc", "", "match response with specified status code (-mc 200,302)"),
		flagSet.StringVarP(&matchLength, "match-length", "ml", "", "match response with specified content length (-ml 100,102)"),
		flagSet.StringVarP(&matchString, "match-string", "ms", "", "match response with specified string (-ms admin)"),
		flagSet.StringVarP(&matchRegex, "match-regex", "mr", "", "match response with specified regex (-mr admin)"),
	)

	var (
		filterStatusCode string
		filterLength     string
		filterString     string
		filterRegex      string
	)
	flagSet.CreateGroup("filters", "Filters",
		flagSet.StringVarP(&filterStatusCode, "filter-code", "fc", "", "filter response with specified status code (-fc 403,401)"),
		flagSet.StringVarP(&filterLength, "filter-length", "fl", "", "filter response with specified content length (-ml 100,102)"),
		flagSet.StringVarP(&filterString, "filter-string", "fs", "", "filter response with specified string (-fs admin)"),
		flagSet.StringVarP(&filterRegex, "filter-regex", "fr", "", "filter response with specified regex (-fe admin)"),
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
