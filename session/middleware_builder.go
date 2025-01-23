// Copyright 2023 ecodeclub
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package session

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/ecodeclub/ginx/gctx"
	"github.com/gin-gonic/gin"
)

// MiddlewareBuilder 登录校验
type MiddlewareBuilder struct {
	sp Provider
	// 当 token 的有效时间少于这个值的时候，就会刷新一下 token
	Threshold time.Duration
}

func (b *MiddlewareBuilder) Build() gin.HandlerFunc {
	threshold := b.Threshold.Milliseconds()
	return func(ctx *gin.Context) {
		ctxx := &gctx.Context{Context: ctx}
		sess, err := b.sp.Get(ctxx)
		if err != nil {
			slog.Debug("未授权", slog.Any("err", err))
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		expiration := sess.Claims().Expiration
		if expiration-time.Now().UnixMilli() < threshold {
			// 刷新一个token
			err = b.sp.RenewAccessToken(ctxx)
			if err != nil {
				slog.Warn("刷新 token 失败", slog.String("err", err.Error()))
			}
		}
		ctx.Set(CtxSessionKey, sess)
	}
}
