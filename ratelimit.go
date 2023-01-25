package main

import (
	"sync"
	"time"
)

type rateLimiter struct {
	sync.Mutex
	delay time.Duration
	ops   map[string]time.Time
	slow  bool
}

func newRateLimiter(delay time.Duration, slow bool) *rateLimiter {
	return &rateLimiter{
		delay: delay,
		ops:   make(map[string]time.Time),
		slow:  slow,
	}
}

// Block 以保证相同host至少间隔设定的delay时间
func (r *rateLimiter) Block(key string) {
	now := time.Now()

	r.Lock()
	// 如果没有map，则直接返回，意味着不限速
	if _, ok := r.ops[key]; !ok {
		r.ops[key] = now
		r.Unlock()
		return
	}

	// ops存有host，需要限速
	t := r.ops[key]
	deadline := t.Add(r.delay)
	if now.After(deadline) { //如果超过了deadline
		r.ops[key] = now
		r.Unlock()
		return
	}
	// 剩余时间
	remaining := deadline.Sub(now)

	// 设置操作时间
	if r.slow {
		r.ops[key] = now.Add(remaining)
	} else {
		r.ops[key] = now
	}
	// time.After(d) 替换为 NewTimer(d).C 意图提高效率
	<-time.NewTimer(remaining).C // 等待剩余时间
	r.Unlock()
}
