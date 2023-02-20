package opt

import (
	"github.com/KagamigawaMeguri/mag/lib"
	"regexp"
	"time"
)

type Options struct {
	Paths            string
	Hosts            string
	Output           string
	DisableOutput    bool
	Method           string
	Headers          lib.CustomHeaders
	RandomAgent      bool
	Body             string
	Threads          int
	Delay            time.Duration
	Timeout          int
	FollowRedirects  bool
	Slow             bool
	Backup           bool
	Proxy            string
	MatchStatusCode  []int
	MatchLength      []int
	MatchString      string
	MatchRegex       *regexp.Regexp
	FilterStatusCode []int
	FilterLength     []int
	FilterString     string
	FilterRegex      *regexp.Regexp
	Verbose          bool
}
