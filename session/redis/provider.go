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
	"time"

	"github.com/ecodeclub/ginx/session/header"

	"github.com/ecodeclub/ginx"

	"github.com/ecodeclub/ginx/gctx"
	ijwt "github.com/ecodeclub/ginx/internal/jwt"
	"github.com/ecodeclub/ginx/session"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var _ session.Provider = &SessionProvider{}

// SessionProvider 默认情况下，产生的 Session 一个 token，
// 而如何返回，以及如何携带，取决于具体的 TokenCarrier 实现
// 很多字段并没有暴露，如果你需要自定义，可以发 issue
type SessionProvider struct {
	client       redis.Cmdable
	m            ijwt.Manager[session.Claims]
	TokenCarrier session.TokenCarrier
	expiration   time.Duration
}

func (rsp *SessionProvider) Destroy(ctx *gctx.Context) error {
	sess, err := rsp.Get(ctx)
	if err != nil {
		return err
	}
	// 清除 token
	rsp.TokenCarrier.Clear(ctx)
	return sess.Destroy(ctx)
}

// UpdateClaims 在这个实现里面，claims 同时写进去了
func (rsp *SessionProvider) UpdateClaims(ctx *gctx.Context, claims session.Claims) error {
	accessToken, err := rsp.m.GenerateAccessToken(claims)
	if err != nil {
		return err
	}
	rsp.TokenCarrier.Inject(ctx, accessToken)
	return nil
}

func (rsp *SessionProvider) RenewAccessToken(ctx *ginx.Context) error {
	// 此时这里应该放着 RefreshToken
	rt := rsp.TokenCarrier.Extract(ctx)
	jwtClaims, err := rsp.m.VerifyAccessToken(rt)
	if err != nil {
		return err
	}
	claims := jwtClaims.Data
	accessToken, err := rsp.m.GenerateAccessToken(claims)
	rsp.TokenCarrier.Inject(ctx, accessToken)
	return err
}

// NewSession 的时候，要先把这个 data 写入到对应的 token 里面
func (rsp *SessionProvider) NewSession(ctx *gctx.Context,
	uid int64,
	jwtData map[string]string,
	sessData map[string]any) (session.Session, error) {
	ssid := uuid.New().String()
	claims := session.Claims{Uid: uid,
		SSID:       ssid,
		Expiration: time.Now().Add(rsp.expiration).UnixMilli(),
		Data:       jwtData}
	accessToken, err := rsp.m.GenerateAccessToken(claims)
	if err != nil {
		return nil, err
	}
	rsp.TokenCarrier.Inject(ctx, accessToken)
	res := newRedisSession(ssid, rsp.expiration, rsp.client, claims)
	if sessData == nil {
		sessData = make(map[string]any, 1)
	}
	sessData["uid"] = uid
	err = res.init(ctx, sessData)
	return res, err
}

// Get 返回 Session，如果没有拿到 session 或者 session 已经过期，会返回 error
func (rsp *SessionProvider) Get(ctx *gctx.Context) (session.Session, error) {
	val, _ := ctx.Get(session.CtxSessionKey)
	// 对接口断言，而不是对实现断言
	res, ok := val.(session.Session)
	if ok {
		return res, nil
	}
	token := rsp.TokenCarrier.Extract(ctx)
	claims, err := rsp.m.VerifyAccessToken(token)
	if err != nil {
		return nil, err
	}
	res = newRedisSession(claims.Data.SSID, rsp.expiration, rsp.client, claims.Data)
	return res, nil
}

// NewSessionProvider 用于管理 Session
func NewSessionProvider(client redis.Cmdable, jwtKey string,
	expiration time.Duration) *SessionProvider {
	// 长 token 过期时间，被看做是 Session 的过期时间
	m := ijwt.NewManagement[session.Claims](ijwt.NewOptions(expiration, jwtKey))
	return &SessionProvider{
		client:       client,
		TokenCarrier: header.NewTokenCarrier(),
		m:            m,
		expiration:   expiration,
	}
}
