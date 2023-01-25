package main

import (
	mapset "github.com/deckarep/golang-set"
	"sync"
)

const (
	lengthThreshold = 5 //相同网页长度数量阈值
)

type SimpleFilter struct {
	UniqueSet mapset.Set
	LengthMap sync.Map
}

func (s *SimpleFilter) DoFilter(req response) bool {
	return s.UniqueFilter(req) || s.LengthFilter(req)
}

// UniqueFilter 请求去重
func (s *SimpleFilter) UniqueFilter(req response) bool {
	if s.UniqueSet == nil {
		s.UniqueSet = mapset.NewSet()
	}
	uid := req.uniqueId()
	if s.UniqueSet.Contains(uid) {
		//存在重复uid，则过滤
		return true
	} else {
		s.UniqueSet.Add(uid)
		return false
	}
}

// LengthFilter 根据重复长度去重
func (s *SimpleFilter) LengthFilter(req response) bool {
	bid := req.bodyId()
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
