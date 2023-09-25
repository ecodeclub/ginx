package ratelimit

import (
	"context"
	_ "embed"
	"time"

	"github.com/redis/go-redis/v9"
)

//go:embed slide_window.lua
var luaSlideWindow string

// RedisSlidingWindowLimiter Redis 上的滑动窗口算法限流器实现
type RedisSlidingWindowLimiter struct {
	Cmd redis.Cmdable

	// 窗口大小
	Interval time.Duration
	// 阈值
	Rate int
	// Interval 内允许 Rate 个请求
	// 1s 内允许 3000 个请求
}

func (r *RedisSlidingWindowLimiter) Limit(ctx context.Context, key string) (bool, error) {
	return r.Cmd.Eval(ctx, luaSlideWindow, []string{key},
		r.Interval.Milliseconds(), r.Rate, time.Now().UnixMilli()).Bool()
}
