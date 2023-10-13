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

package jwt

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/ecodeclub/ekit/set"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// MiddlewareBuilder 创建一个校验登录的 middleware
// ignorePath: 默认使用 func(path string) bool { return false } 也就是全部不忽略.
type MiddlewareBuilder[T any] struct {
	ignorePath func(path string) bool // Middleware 方法中忽略认证的路径
	manager    *Management[T]
	nowFunc    func() time.Time // 控制 jwt 的时间
}

func newMiddlewareBuilder[T any](m *Management[T]) *MiddlewareBuilder[T] {
	return &MiddlewareBuilder[T]{
		manager: m,
		ignorePath: func(path string) bool {
			return false
		},
		nowFunc: m.nowFunc,
	}
}

func (m *MiddlewareBuilder[T]) IgnorePath(path ...string) *MiddlewareBuilder[T] {
	return m.IgnorePathFunc(staticIgnorePaths(path...))
}

// IgnorePathFunc 设置忽略资源令牌认证的路径.
func (m *MiddlewareBuilder[T]) IgnorePathFunc(fn func(path string) bool) *MiddlewareBuilder[T] {
	m.ignorePath = fn
	return m
}

func (m *MiddlewareBuilder[T]) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 不需要校验
		if m.ignorePath(ctx.Request.URL.Path) {
			return
		}

		// 提取 token
		tokenStr := m.manager.extractTokenString(ctx)
		if tokenStr == "" {
			slog.Debug("failed to extract token")
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 校验 token
		clm, err := m.manager.VerifyAccessToken(tokenStr,
			jwt.WithTimeFunc(m.nowFunc))
		if err != nil {
			slog.Debug("access token verification failed")
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 设置 claims
		m.manager.SetClaims(ctx, clm)
	}
}

// staticIgnorePaths 设置静态忽略的路径.
func staticIgnorePaths(paths ...string) func(path string) bool {
	s := set.NewMapSet[string](len(paths))
	for _, path := range paths {
		s.Add(path)
	}
	return func(path string) bool {
		return s.Exist(path)
	}
}
