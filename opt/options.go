package opt

import (
	"github.com/KagamigawaMeguri/mag/lib"
	mapset "github.com/deckarep/golang-set/v2"
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
	Proxy            string
	MatchStatusCode  mapset.Set[int]
	MatchLength      mapset.Set[int]
	MatchString      string
	MatchRegex       *regexp.Regexp
	FilterStatusCode mapset.Set[int]
	FilterLength     mapset.Set[int]
	FilterString     string
	FilterRegex      *regexp.Regexp
	Verbose          bool
}
