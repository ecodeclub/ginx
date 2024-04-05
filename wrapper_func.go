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

package ginx

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/ecodeclub/ginx/session"
	"github.com/gin-gonic/gin"
)

func W(fn func(ctx *Context) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		res, err := fn(&Context{Context: ctx})
		if errors.Is(err, ErrNoResponse) {
			slog.Debug("不需要响应", slog.Any("err", err))
			return
		}
		if errors.Is(err, ErrUnauthorized) {
			slog.Debug("未授权", slog.Any("err", err))
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if err != nil {
			slog.Error("执行业务逻辑失败", slog.Any("err", err))
			ctx.PureJSON(http.StatusInternalServerError, res)
			return
		}
		ctx.PureJSON(http.StatusOK, res)
	}
}

func B[Req any](fn func(ctx *Context, req Req) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req Req
		if err := ctx.Bind(&req); err != nil {
			slog.Debug("绑定参数失败", slog.Any("err", err))
			return
		}
		res, err := fn(&Context{Context: ctx}, req)
		if errors.Is(err, ErrNoResponse) {
			slog.Debug("不需要响应", slog.Any("err", err))
			return
		}
		if errors.Is(err, ErrUnauthorized) {
			slog.Debug("未授权", slog.Any("err", err))
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if err != nil {
			slog.Error("执行业务逻辑失败", slog.Any("err", err))
			ctx.PureJSON(http.StatusInternalServerError, res)
			return
		}
		ctx.PureJSON(http.StatusOK, res)
	}
}

// BS 的意思是，传入的业务逻辑方法可以接受 req 和 sess 两个参数
func BS[Req any](fn func(ctx *Context, req Req, sess session.Session) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		gtx := &Context{Context: ctx}
		sess, err := session.Get(gtx)
		if err != nil {
			slog.Debug("获取 Session 失败", slog.Any("err", err))
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		var req Req
		// Bind 方法本身会返回 400 的错误
		if err := ctx.Bind(&req); err != nil {
			slog.Debug("绑定参数失败", slog.Any("err", err))
			return
		}
		res, err := fn(gtx, req, sess)
		if errors.Is(err, ErrNoResponse) {
			slog.Debug("不需要响应", slog.Any("err", err))
			return
		}
		// 如果里面有权限校验，那么会返回 401 错误（目前来看，主要是登录态校验）
		if errors.Is(err, ErrUnauthorized) {
			slog.Debug("未授权", slog.Any("err", err))
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if err != nil {
			slog.Error("执行业务逻辑失败", slog.Any("err", err))
			ctx.PureJSON(http.StatusInternalServerError, res)
			return
		}
		ctx.PureJSON(http.StatusOK, res)
	}
}

// S 的意思是，传入的业务逻辑方法可以接受 Session 参数
func S(fn func(ctx *Context, sess session.Session) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		gtx := &Context{Context: ctx}
		sess, err := session.Get(gtx)
		if err != nil {
			slog.Debug("获取 Session 失败", slog.Any("err", err))
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		res, err := fn(gtx, sess)
		if errors.Is(err, ErrNoResponse) {
			slog.Debug("不需要响应", slog.Any("err", err))
			return
		}
		// 如果里面有权限校验，那么会返回 401 错误（目前来看，主要是登录态校验）
		if errors.Is(err, ErrUnauthorized) {
			slog.Debug("未授权", slog.Any("err", err))
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if err != nil {
			slog.Error("执行业务逻辑失败", slog.Any("err", err))
			ctx.PureJSON(http.StatusInternalServerError, res)
			return
		}
		ctx.PureJSON(http.StatusOK, res)
	}
}
