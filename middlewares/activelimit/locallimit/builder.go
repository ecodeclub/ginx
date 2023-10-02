package locallimit

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/atomic"
	"net/http"
)

type LocalActiveLimit struct {
	//最大限制个数
	maxActive *atomic.Int64
	//当前活跃个数
	countActive *atomic.Int64
}

// NewLocalActiveLimit 全局限流
func NewLocalActiveLimit(maxActive int64) *LocalActiveLimit {
	return &LocalActiveLimit{
		maxActive:   atomic.NewInt64(maxActive),
		countActive: atomic.NewInt64(0),
	}
}

func (a *LocalActiveLimit) SetMaxActive(maxActive int64) *LocalActiveLimit {
	a.maxActive.Store(maxActive)
	return a
}

func (a *LocalActiveLimit) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		current := a.countActive.Add(1)
		defer func() {
			a.countActive.Sub(1)
		}()
		if current <= a.maxActive.Load() {
			ctx.Next()
		} else {
			ctx.AbortWithStatus(http.StatusTooManyRequests)
		}
		return

	}
}
