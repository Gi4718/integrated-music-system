package util

import (
	"sync"
	"time"
)

// RateLimiter 自适应速率限制器
type RateLimiter struct {
	mu           sync.Mutex
	lastAccess   time.Time
	interval     time.Duration
	minInterval  time.Duration
	maxInterval  time.Duration
	cooldown     int // 连续成功次数，用于逐步降低间隔
}

// NewRateLimiter 创建自适应速率限制器
func NewRateLimiter(minInterval time.Duration) *RateLimiter {
	maxInterval := minInterval * 10
	if maxInterval < 10*time.Second {
		maxInterval = 10 * time.Second
	}
	return &RateLimiter{
		interval:    minInterval,
		minInterval: minInterval,
		maxInterval: maxInterval,
	}
}

// Wait 等待直到可以执行
func (r *RateLimiter) Wait() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	if !r.lastAccess.IsZero() {
		elapsed := now.Sub(r.lastAccess)
		if elapsed < r.interval {
			time.Sleep(r.interval - elapsed)
		}
	}
	r.lastAccess = time.Now()
}

// Increase 遇到限速时增大间隔
func (r *RateLimiter) Increase() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.interval = r.interval * 2
	if r.interval > r.maxInterval {
		r.interval = r.maxInterval
	}
	r.cooldown = 0
}

// Decrease 连续成功后逐步缩小间隔
func (r *RateLimiter) Decrease() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cooldown++
	if r.cooldown >= 20 && r.interval > r.minInterval {
		r.interval = r.interval / 2
		if r.interval < r.minInterval {
			r.interval = r.minInterval
		}
		r.cooldown = 0
	}
}

// GetInterval 获取当前间隔（用于日志）
func (r *RateLimiter) GetInterval() time.Duration {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.interval
}

// RetryConfig 重试配置
type RetryConfig struct {
	MaxAttempts int
	InitialWait time.Duration
	MaxWait     time.Duration
	Multiplier  float64
}

// DefaultRetryConfig 默认重试配置
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 3,
		InitialWait: 1 * time.Second,
		MaxWait:     30 * time.Second,
		Multiplier:  2.0,
	}
}

// RetryWithBackoff 带指数退避的重试
func RetryWithBackoff(fn func() error, config RetryConfig) error {
	var err error
	wait := config.InitialWait

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		err = fn()
		if err == nil {
			return nil
		}

		if attempt < config.MaxAttempts-1 {
			time.Sleep(wait)
			wait = time.Duration(float64(wait) * config.Multiplier)
			if wait > config.MaxWait {
				wait = config.MaxWait
			}
		}
	}

	return err
}
