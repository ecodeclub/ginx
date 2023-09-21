package ratelimit

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/ecodeclub/ginx/internal/ratelimit"
)

type Builder struct {
	limiter  ratelimit.Limiter
	genKeyFn func(ctx *gin.Context) string
	logFn    func(msg any, args ...any)
}

// NewBuilder
// genKeyFn: 默认使用 IP 限流.
// logFn: 默认使用 log.Println().
func NewBuilder(limiter ratelimit.Limiter) *Builder {
	return &Builder{
		limiter: limiter,
		genKeyFn: func(ctx *gin.Context) string {
			var b strings.Builder
			b.WriteString("ip-limiter")
			b.WriteString(":")
			b.WriteString(ctx.ClientIP())
			return b.String()
		},
		logFn: func(msg any, args ...any) {
			v := make([]any, 0, len(args)+1)
			v = append(v, msg)
			v = append(v, args...)
			log.Println(v...)
		},
	}
}

func (b *Builder) SetKeyGenFunc(fn func(*gin.Context) string) *Builder {
	b.genKeyFn = fn
	return b
}

func (b *Builder) SetLogFunc(fn func(msg any, args ...any)) *Builder {
	b.logFn = fn
	return b
}

func (b *Builder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		limited, err := b.limit(ctx)
		if err != nil {
			b.logFn(err)
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if limited {
			ctx.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		ctx.Next()
	}
}

func (b *Builder) limit(ctx *gin.Context) (bool, error) {
	return b.limiter.Limit(ctx, b.genKeyFn(ctx))
}
