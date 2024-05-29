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
	"errors"
	"strings"
	"time"

	"github.com/ecodeclub/ginx"

	"github.com/ecodeclub/ginx/gctx"
	ijwt "github.com/ecodeclub/ginx/internal/jwt"
	"github.com/ecodeclub/ginx/session"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var (
	keyRefreshToken = "refresh_token"
)

var _ session.Provider = &SessionProvider{}

// SessionProvider 默认情况下，产生的 Session 对应了两个 token，
// access token 和 refresh token
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

// UpdateClaims 在这个实现里面，claims 同时写进去了
func (rsp *SessionProvider) UpdateClaims(ctx *gctx.Context, claims session.Claims) error {
	accessToken, err := rsp.m.GenerateAccessToken(claims)
	if err != nil {
		return err
	}
	refreshToken, err := rsp.m.GenerateRefreshToken(claims)
	if err != nil {
		return err
	}
	ctx.Header(rsp.atHeader, accessToken)
	ctx.Header(rsp.rtHeader, refreshToken)
	return nil
}

func (rsp *SessionProvider) RenewAccessToken(ctx *ginx.Context) error {
	// 此时这里应该放着 RefreshToken
	rt := rsp.extractTokenString(ctx)
	jwtClaims, err := rsp.m.VerifyRefreshToken(rt)
	if err != nil {
		return err
	}
	claims := jwtClaims.Data
	sess := newRedisSession(claims.SSID, rsp.expiration, rsp.client, claims)
	oldToken := sess.Get(ctx, keyRefreshToken).StringOrDefault("")
	// refresh_token 只能用一次，不管成功与否
	_ = sess.Del(ctx, keyRefreshToken)
	// 说明这个 rt 是已经用过的 refreshToken
	// 或者 session 本身就已经过期了
	if oldToken != rt {
		return errors.New("refresh_token 已经过期")
	}
	accessToken, err := rsp.m.GenerateAccessToken(claims)
	if err != nil {
		return err
	}
	refreshToken, err := rsp.m.GenerateRefreshToken(claims)
	if err != nil {
		return err
	}
	ctx.Header(rsp.rtHeader, refreshToken)
	ctx.Header(rsp.atHeader, accessToken)
	return sess.Set(ctx, keyRefreshToken, refreshToken)
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
	sessData[keyRefreshToken] = refreshToken
	err = res.init(ctx, sessData)
	return res, err
}

// extractTokenString 提取 token 字符串.
func (rsp *SessionProvider) extractTokenString(ctx *ginx.Context) string {
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
	// 对接口断言，而不是对实现断言
	res, ok := val.(session.Session)
	if ok {
		return res, nil
	}

	claims, err := rsp.m.VerifyAccessToken(rsp.extractTokenString(ctx))
	if err != nil {
		return nil, err
	}
	res = newRedisSession(claims.Data.SSID, rsp.expiration, rsp.client, claims.Data)
	return res, nil
}

// NewSessionProvider 长短 token + session 机制。短 token 的过期时间是一小时
// 长 token 的过期时间是 30 天
func NewSessionProvider(client redis.Cmdable, jwtKey string) *SessionProvider {
	// 长 token 过期时间，被看做是 Session 的过期时间
	expiration := time.Hour * 24 * 30
	m := ijwt.NewManagement[session.Claims](ijwt.NewOptions(time.Hour, jwtKey),
		ijwt.WithRefreshJWTOptions[session.Claims](ijwt.NewOptions(expiration, jwtKey)))
	return &SessionProvider{
		client:      client,
		atHeader:    "X-Access-Token",
		rtHeader:    "X-Refresh-Token",
		tokenHeader: "Authorization",
		m:           m,
		expiration:  expiration,
	}
}
