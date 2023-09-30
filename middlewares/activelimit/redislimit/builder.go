package redislimit

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/atomic"
	"net/http"
)

type RedisActiveLimit struct {
	//最大限制个数
	MaxActive *atomic.Int64

	//是否开启限流
	Statue *atomic.Bool
	//用来记录连接数的key
	key string
	cmd redis.Cmdable
}

// NewRedisActiveLimit 全局限流
func NewRedisActiveLimit(cmd redis.Cmdable, maxActive int64, key string) *RedisActiveLimit {
	return &RedisActiveLimit{
		MaxActive: atomic.NewInt64(maxActive),
		Statue:    atomic.NewBool(true),
		key:       key,
		cmd:       cmd,
	}
}

func (a *RedisActiveLimit) SetStatue(statue bool) *RedisActiveLimit {
	a.Statue.Store(statue)
	return a
}

func (a *RedisActiveLimit) SetMaxActive(maxActive int64) *RedisActiveLimit {
	a.MaxActive.Store(maxActive)
	return a
}

func (a *RedisActiveLimit) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//开启限流
		if a.Statue.Load() {
			currentCount, err := a.cmd.Incr(ctx, a.key).Result()
			if err != nil {
				//为了安全性 直接返回异常
				ctx.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			defer func() {
				_ = a.cmd.Decr(ctx, a.key).Err()
			}()
			if currentCount <= a.MaxActive.Load() {
				ctx.Next()
			} else {
				ctx.AbortWithStatus(http.StatusTooManyRequests)
			}
			return

		}
		ctx.Next()
	}
}
