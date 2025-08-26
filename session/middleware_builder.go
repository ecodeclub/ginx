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
	"sync"
	"time"

	"github.com/ecodeclub/ginx/gctx"
	"github.com/gin-gonic/gin"
)

// MiddlewareOption 定义 middleware 的配置选项
type MiddlewareOption func(*MiddlewareBuilder)

// WithThreshold 设置续期阈值
func WithThreshold(threshold time.Duration) MiddlewareOption {
	return func(b *MiddlewareBuilder) {
		b.threshold = threshold
	}
}

// WithConcurrencyControl 启用并发控制
func WithConcurrencyControl() MiddlewareOption {
	return func(b *MiddlewareBuilder) {
		b.enableConcurrencyControl = true
	}
}

// WithLogger 设置自定义日志器
func WithLogger(logger *slog.Logger) MiddlewareOption {
	return func(b *MiddlewareBuilder) {
		b.logger = logger
	}
}

// MiddlewareBuilder 登录校验
type MiddlewareBuilder struct {
	sp Provider
	// 当 token 的有效时间少于这个值的时候，就会刷新一下 token
	threshold time.Duration
	logger    *slog.Logger

	// 用于并发控制的锁
	enableConcurrencyControl bool
	renewalLocks             sync.Map
}

func NewMiddlewareBuilder(sp Provider, opts ...MiddlewareOption) *MiddlewareBuilder {
	builder := &MiddlewareBuilder{
		sp:                       sp,
		threshold:                time.Minute * 30,
		enableConcurrencyControl: false,
		logger:                   slog.Default(),
	}

	// 应用选项
	for _, opt := range opts {
		opt(builder)
	}

	return builder
}

func (b *MiddlewareBuilder) Build() gin.HandlerFunc {
	threshold := b.threshold.Milliseconds()
	return func(ctx *gin.Context) {
		gCtx := &gctx.Context{Context: ctx}
		sess, err := b.sp.Get(gCtx)
		if err != nil {
			b.logger.Debug("未授权", slog.Any("err", err))
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		expiration := sess.Claims().Expiration
		if expiration-time.Now().UnixMilli() < threshold {
			// 如果需要并发控制，使用锁机制
			if b.enableConcurrencyControl {
				b.renewWithConcurrencyControl(gCtx, sess.Claims().SSID)
			} else {
				// 直接续期
				err = b.sp.RenewAccessToken(gCtx)
				if err != nil {
					b.logger.Warn("刷新 token 失败", slog.String("err", err.Error()))
				}
			}
		}
		ctx.Set(CtxSessionKey, sess)
	}
}

// renewWithConcurrencyControl 使用并发控制进行续期
func (b *MiddlewareBuilder) renewWithConcurrencyControl(ctx *gctx.Context, ssid string) {
	// 获取或创建该 SSID 的锁
	lockInterface, _ := b.renewalLocks.LoadOrStore(ssid, &sync.Mutex{})
	lock := lockInterface.(*sync.Mutex)

	// 尝试获取锁
	if lock.TryLock() {
		defer lock.Unlock()

		// 执行续期
		err := b.sp.RenewAccessToken(ctx)
		if err != nil {
			b.logger.Warn("刷新 token 失败", slog.String("err", err.Error()))
		}
	} else {
		// 如果获取不到锁，说明其他请求正在处理续期，跳过本次续期
		b.logger.Debug("跳过续期，其他请求正在处理", slog.String("ssid", ssid))
	}
}
