package ratelimiter

import (
	"sync"
)

type IPRateLimiter struct {
	limiters map[string]*RateLimiter
	mutex    sync.Mutex
}

func NewIpRateLimiter() *IPRateLimiter {
	return &IPRateLimiter{
		limiters: make(map[string]*RateLimiter),
	}
}

func (i *IPRateLimiter) GetLimiter(ip string) *RateLimiter {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	limiter, exists := i.limiters[ip]
	if !exists {
		// 240 rpm (burst of 240 tokens, refill 4 tokens/sec)
		limiter = NewRateLimiter(240, 4)
		i.limiters[ip] = limiter
	}

	return limiter
}
