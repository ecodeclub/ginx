package session

import (
	"log/slog"
	"net/http"

	"github.com/ecodeclub/ginx/gctx"
	"github.com/gin-gonic/gin"
)

// MiddlewareBuilder 登录校验
type MiddlewareBuilder struct {
	sp Provider
}

func (b *MiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		sess, err := b.sp.Get(&gctx.Context{Context: ctx})
		if err != nil {
			slog.Debug("未授权", slog.Any("err", err))
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		ctx.Set(CtxSessionKey, sess)
	}
}
