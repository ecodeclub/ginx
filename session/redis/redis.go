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

package redis

import (
	"context"
	"strings"
	"time"

	"github.com/ecodeclub/ekit"
	"github.com/ecodeclub/ginx/gctx"
	"github.com/ecodeclub/ginx/internal/errs"
	ijwt "github.com/ecodeclub/ginx/internal/jwt"
	"github.com/ecodeclub/ginx/session"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Session 生命周期应该和 http 请求保持一致
// 注意该实现本身预加载了 Session 的所有数据
type Session struct {
	client redis.Cmdable
	// key 是 ssid 拼接而成。注意，它不是 access token，也不是 refresh token
	key        string
	data       map[string]string
	claims     session.Claims
	expiration time.Duration
}

func newRedisSession(
	ssid string,
	expiration time.Duration,
	client redis.Cmdable, cl session.Claims) *Session {
	return &Session{
		client:     client,
		key:        "session:" + ssid,
		expiration: expiration,
		claims:     cl,
	}
}

func (r *Session) Set(ctx context.Context, key string, val any) error {
	return r.client.HMSet(ctx, r.key, key, val).Err()
}

func (r *Session) init(ctx context.Context, kvs map[string]any) error {
	pip := r.client.Pipeline()
	for k, v := range kvs {
		pip.HMSet(ctx, r.key, k, v)
	}
	pip.Expire(ctx, r.key, r.expiration)
	_, err := pip.Exec(ctx)
	return err
}

func (r *Session) Get(ctx context.Context, key string) ekit.AnyValue {
	val, ok := r.data[key]
	if ok {
		return ekit.AnyValue{Val: val}
	}
	return ekit.AnyValue{
		// 报错
		Err: errs.ErrSessionKeyNotFound,
	}
}

func (r *Session) preload(ctx context.Context) error {
	var err error
	r.data, err = r.client.HGetAll(ctx, r.key).Result()
	if err != nil {
		return err
	}
	if len(r.data) == 0 {
		return errs.ErrUnauthorized
	}
	return nil
}

func (r *Session) Claims() session.Claims {
	return r.claims
}

// SessionProvider 默认是预加载机制，即 Get 的时候会顺便把所有的数据都拿过来
// 默认情况下，产生的 Session 对应了两个 token，access token 和 refresh token
// 它们会被放进去 http.Response x-access-token 和 x-refresh-token 里面
// 后续前端发送请求的时候，需要把 token 放到 Authorization 中，以 Bearer 的形式传过来
// 很多字段并没有暴露，如果你需要自定义，可以发 issue
type SessionProvider struct {
	client      redis.Cmdable
	m           ijwt.Manager[session.Claims]
	tokenHeader string // 认证的请求头(存放 token 的请求头 key)
	atHeader    string // 暴露到外部的资源请求头
	rtHeader    string // 暴露到外部的刷新请求头
	// 这个是长 token 的过期时间
	expiration time.Duration
}

// NewSessionProvider 长短 token + session 机制。短 token 的过期时间是一小时
// 长 token 的过期时间是 30 天
func NewSessionProvider(client redis.Cmdable, key string) *SessionProvider {
	// 长 token 过期时间，被看做是 Session 的过期时间
	expiration := time.Hour * 24 * 30
	m := ijwt.NewManagement[session.Claims](ijwt.NewOptions(time.Hour, key),
		ijwt.WithRefreshJWTOptions[session.Claims](ijwt.NewOptions(expiration, key)))
	return &SessionProvider{
		client:      client,
		atHeader:    "X-Access-Token",
		rtHeader:    "X-Refresh-Token",
		tokenHeader: "Authorization",
		m:           m,
		expiration:  expiration,
	}
}

// NewSession 的时候，要先把这个 data 写入到对应的 token 里面
func (rsp *SessionProvider) NewSession(ctx *gctx.Context,
	uid int64,
	jwtData map[string]string,
	sessData map[string]any) (session.Session, error) {
	ssid := uuid.New().String()
	claims := session.Claims{Uid: uid, SSID: ssid, Data: jwtData}
	accessToken, err := rsp.m.GenerateAccessToken(claims)
	if err != nil {
		return nil, err
	}
	refreshToken, err := rsp.m.GenerateRefreshToken(claims)
	if err != nil {
		return nil, err
	}

	ctx.Header(rsp.rtHeader, refreshToken)
	ctx.Header(rsp.atHeader, accessToken)

	res := newRedisSession(ssid, rsp.expiration, rsp.client, claims)
	// 将 refresh token 放进去 redis 里面
	// refresh token 应该只能用一次
	// 要设置超时时间
	if sessData == nil {
		sessData = make(map[string]any, 2)
	}
	sessData["uid"] = uid
	sessData["refresh_token"] = refreshToken
	err = res.init(ctx, sessData)
	return res, err
}

// extractTokenString 提取 token 字符串.
func (rsp *SessionProvider) extractTokenString(ctx *gin.Context) string {
	authCode := ctx.GetHeader(rsp.tokenHeader)
	const bearerPrefix = "Bearer "
	if strings.HasPrefix(authCode, bearerPrefix) {
		return authCode[len(bearerPrefix):]
	}
	return ""
}

// Get 返回 Session，如果没有拿到 session 或者 session 已经过期，会返回 error
func (rsp *SessionProvider) Get(ctx *gctx.Context) (session.Session, error) {
	val, _ := ctx.Get(session.CtxSessionKey)
	res, ok := val.(*Session)
	if ok {
		return res, nil
	}

	claims, err := rsp.m.VerifyAccessToken(rsp.extractTokenString(ctx.Context))
	if err != nil {
		return nil, err
	}
	res = newRedisSession(claims.Data.SSID, rsp.expiration, rsp.client, claims.Data)
	return res, res.preload(ctx)
}
