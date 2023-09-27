package activelimit

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/atomic"

	"github.com/ecodeclub/ginx/internal/activelimit"
)

type ActiveLimit struct {
	limiter activelimit.Limiter
	//最大限制个数
	MaxActive *atomic.Int64
	////当前活跃个数
	//CountActive *atomic.Int64
	//是否开启限流
	Statue   *atomic.Bool
	genKeyFn func(ctx *gin.Context) string
}

// NewActiveLimit 全局限流
func NewActiveLimit(maxActive int64, limiter activelimit.Limiter, genKeyFn func(ctx *gin.Context) string) *ActiveLimit {
	return &ActiveLimit{
		limiter:   limiter,
		MaxActive: atomic.NewInt64(maxActive),
		//CountActive: atomic.NewInt64(0),
		Statue:   atomic.NewBool(true),
		genKeyFn: genKeyFn,
	}
}

func (b *ActiveLimit) SetStatue(statue bool) *ActiveLimit {
	b.Statue.Store(statue)
	return b
}

func (b *ActiveLimit) SetMaxActive(maxActive int64) *ActiveLimit {
	b.MaxActive.Store(maxActive)
	return b
}

func (b *ActiveLimit) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//开启限流
		if b.Statue.Load() {
			////直接减一操作 下面必加一成功
			//defer b.countActive.Sub(1)
			//// 并直接占坑成功
			//if b.countActive.Add(1) <= b.MaxActive.Load() {
			//	ctx.Next()
			//} else {
			//	ctx.AbortWithStatus(http.StatusTooManyRequests)
			//}
			//return
			limited, err := b.limiter.Add(ctx, b.genKeyFn(ctx), b.MaxActive.Load())
			if err != nil {
				ctx.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			if limited {
				ctx.AbortWithStatus(http.StatusTooManyRequests)
			} else {
				defer func() {
					limited, err = b.limiter.Sub(ctx, b.genKeyFn(ctx))

					if err != nil {
						fmt.Println("减一操作异常,需要关注一下:", err)
						return
					}
					if !limited {
						fmt.Println("可能有人搞你的redis 或者代码有问题")
					}
				}()
				ctx.Next()
			}
			return

		}
		ctx.Next()
	}
}
