package locallimit

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/atomic"
	"net/http"
)

type LocalActiveLimit struct {
	//最大限制个数
	MaxActive *atomic.Int64
	//当前活跃个数
	countActive *atomic.Int64
	//是否开启限流
	Statue *atomic.Bool
}

// NewLocalActiveLimit 全局限流
func NewLocalActiveLimit(maxActive int64) *LocalActiveLimit {
	return &LocalActiveLimit{
		MaxActive:   atomic.NewInt64(maxActive),
		countActive: atomic.NewInt64(0),
		Statue:      atomic.NewBool(true),
	}
}

func (a *LocalActiveLimit) SetStatue(statue bool) *LocalActiveLimit {
	a.Statue.Store(statue)
	return a
}

func (a *LocalActiveLimit) SetMaxActive(maxActive int64) *LocalActiveLimit {
	a.MaxActive.Store(maxActive)
	return a
}

func (a *LocalActiveLimit) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//开启限流
		if a.Statue.Load() {
			//直接减一操作 下面必加一成功
			defer func() {
				fmt.Println("简易操作....:")
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
		ctx.Next()
	}
}
