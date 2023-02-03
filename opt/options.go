package opt

import (
	"github.com/KagamigawaMeguri/mag/gohttp"
	mapset "github.com/deckarep/golang-set"
	"regexp"
	"time"
)

type Options struct {
	Paths            string
	Hosts            string
	Output           string
	DisableOutput    bool
	Method           string
	Headers          gohttp.CustomHeaders
	RandomAgent      bool
	Body             string
	Threads          int
	Delay            time.Duration
	Timeout          int
	FollowRedirects  bool
	Slow             bool
	Proxy            string
	MatchStatusCode  mapset.Set
	MatchLength      mapset.Set
	MatchString      string
	MatchRegex       *regexp.Regexp
	FilterStatusCode mapset.Set
	FilterLength     mapset.Set
	FilterString     string
	FilterRegex      *regexp.Regexp
	Verbose          bool
}
