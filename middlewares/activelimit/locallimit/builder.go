package locallimit

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/atomic"
	"net/http"
)

type LocalActiveLimit struct {
	//最大限制个数
	MaxActive *atomic.Int64
	//当前活跃个数
	countActive *atomic.Int64
}

// NewLocalActiveLimit 全局限流
func NewLocalActiveLimit(maxActive int64) *LocalActiveLimit {
	return &LocalActiveLimit{
		MaxActive:   atomic.NewInt64(maxActive),
		countActive: atomic.NewInt64(0),
	}
}

func (a *LocalActiveLimit) SetMaxActive(maxActive int64) *LocalActiveLimit {
	a.MaxActive.Store(maxActive)
	return a
}

func (a *LocalActiveLimit) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//直接减一操作 下面必加一成功
		defer func() {
			a.countActive.Sub(1)
		}()
		// 并直接占坑成功
		if a.countActive.Add(1) <= a.MaxActive.Load() {
			ctx.Next()
		} else {
			ctx.AbortWithStatus(http.StatusTooManyRequests)
		}
		return

	}
}
