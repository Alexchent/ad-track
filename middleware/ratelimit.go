package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// TokenBucket 令牌桶结构
type TokenBucket struct {
	capacity     int64      // 桶的容量
	tokens       int64      // 当前令牌数量
	rate         int64      // 令牌生成速率（每秒生成的令牌数）
	lastTime     time.Time  // 上次更新时间
	mu           sync.Mutex // 互斥锁
	refillPeriod int64      // 令牌填充周期（纳秒）
}

// NewTokenBucket 创建一个新的令牌桶
// rate: 每秒生成的令牌数
// capacity: 桶的最大容量
func NewTokenBucket(rate, capacity int64) *TokenBucket {
	return &TokenBucket{
		capacity:     capacity,
		tokens:       capacity, // 初始时桶是满的
		rate:         rate,
		lastTime:     time.Now(),
		refillPeriod: int64(time.Second) / rate, // 计算每个令牌的生成周期
	}
}

// Allow 尝试从桶中获取一个令牌
// 返回 true 表示获取成功，false 表示桶中没有足够的令牌
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	// 计算自上次更新以来应该添加的令牌数量
	elapsed := now.Sub(tb.lastTime).Nanoseconds()
	tokensToAdd := elapsed / tb.refillPeriod

	if tokensToAdd > 0 {
		tb.tokens += tokensToAdd
		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity
		}
		tb.lastTime = now
	}

	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}

// RateLimiter 限流器，支持基于IP的限流
type RateLimiter struct {
	buckets sync.Map // 存储不同IP的令牌桶
	rate    int64    // 每秒生成的令牌数
	cap     int64    // 桶的容量
}

// NewRateLimiter 创建一个新的限流器
// rate: 每秒允许的请求数
// capacity: 桶的容量（允许的突发流量）
func NewRateLimiter(rate, capacity int64) *RateLimiter {
	limiter := &RateLimiter{
		rate: rate,
		cap:  capacity,
	}

	// 启动后台清理协程，定期清理长时间未使用的桶
	go limiter.cleanup()

	return limiter
}

// getBucket 获取或创建指定IP的令牌桶
func (rl *RateLimiter) getBucket(key string) *TokenBucket {
	if bucket, ok := rl.buckets.Load(key); ok {
		return bucket.(*TokenBucket)
	}

	// 创建新的令牌桶
	bucket := NewTokenBucket(rl.rate, rl.cap)
	actual, _ := rl.buckets.LoadOrStore(key, bucket)
	return actual.(*TokenBucket)
}

// Allow 检查指定key是否允许请求
func (rl *RateLimiter) Allow(key string) bool {
	return rl.getBucket(key).Allow()
}

// cleanup 定期清理长时间未使用的令牌桶
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		rl.buckets.Range(func(key, value interface{}) bool {
			bucket := value.(*TokenBucket)
			bucket.mu.Lock()
			// 如果超过5分钟没有请求，删除该桶
			if time.Since(bucket.lastTime) > 5*time.Minute {
				rl.buckets.Delete(key)
			}
			bucket.mu.Unlock()
			return true
		})
	}
}

// 全局限流器实例
var globalLimiter *RateLimiter

// InitRateLimiter 初始化全局限流器
// rate: 每秒允许的请求数
// capacity: 桶的容量（允许的突发流量）
func InitRateLimiter(rate, capacity int64) {
	globalLimiter = NewRateLimiter(rate, capacity)
}

// RateLimit 令牌桶限流中间件
// 使用客户端IP作为限流的key
func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if globalLimiter == nil {
			c.Next()
			return
		}

		// 使用客户端IP作为限流的key
		key := c.ClientIP()

		if !globalLimiter.Allow(key) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "请求过于频繁，请稍后再试",
				"code":  429,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitWithKey 自定义key的限流中间件
// keyFunc: 用于生成限流key的函数
func RateLimitWithKey(keyFunc func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if globalLimiter == nil {
			c.Next()
			return
		}

		key := keyFunc(c)

		if !globalLimiter.Allow(key) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "请求过于频繁，请稍后再试",
				"code":  429,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
