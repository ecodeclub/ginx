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
