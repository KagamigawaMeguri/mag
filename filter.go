package main

import (
	"github.com/KagamigawaMeguri/mag/gohttp"
	mapset "github.com/deckarep/golang-set"
	"github.com/go-dedup/simhash"
	"github.com/go-dedup/simhash/sho"
	"sync"
)

const (
	lengthThreshold = 5 //相同网页长度数量阈值
)

type SimpleFilter struct {
	UniqueSet mapset.Set
	LengthMap sync.Map
	oracle    *sho.Oracle
	sh        *simhash.SimhashBase
	scope     uint8
}

func (s *SimpleFilter) DoFilter(resp *gohttp.Response) bool {
	return s.UniqueFilter(resp) || s.SimhashFilter(resp) || s.LengthFilter(resp)
}

// UniqueFilter 请求去重
func (s *SimpleFilter) UniqueFilter(resp *gohttp.Response) bool {
	uid := resp.UniqueId()
	if s.UniqueSet.Contains(uid) {
		//存在重复uid，则过滤
		return true
	} else {
		s.UniqueSet.Add(uid)
		return false
	}
}

// LengthFilter 根据重复长度去重
func (s *SimpleFilter) LengthFilter(resp *gohttp.Response) bool {
	bid := resp.BodyId()
	var v interface{}
	v, ok := s.LengthMap.Load(bid)
	if ok {
		// 存在
		vint := v.(int)
		vint += 1
		if vint > lengthThreshold {
			//存在并且数量大于阈值，则过滤
			return true
		}
		s.LengthMap.Store(bid, vint)
		return false
	} else {
		s.LengthMap.LoadOrStore(bid, 0)
		return false
	}
}

func (s *SimpleFilter) SimhashFilter(resp *gohttp.Response) bool {
	hash := s.sh.GetSimhash(s.sh.NewWordFeatureSet(resp.Body))
	if s.oracle.Seen(hash, s.scope) {
		return true
	} else {
		s.oracle.See(hash)
		return false
	}
}

func NewSimpleFilter(scope uint8) *SimpleFilter {
	return &SimpleFilter{
		UniqueSet: mapset.NewSet(),
		oracle:    sho.NewOracle(),
		sh:        simhash.NewSimhash(),
		scope:     scope,
	}
}
