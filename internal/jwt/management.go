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
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/ecodeclub/ekit/bean/option"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const bearerPrefix = "Bearer"

var (
	errEmptyRefreshOpts = errors.New("refreshJWTOptions are nil")
)

type Management[T any] struct {
	allowTokenHeader    string // 认证的请求头(存放 token 的请求头 key)
	exposeAccessHeader  string // 暴露到外部的资源请求头
	exposeRefreshHeader string // 暴露到外部的刷新请求头

	accessJWTOptions   Options          // 资源 token 选项
	refreshJWTOptions  *Options         // 刷新 token 选项
	rotateRefreshToken bool             // 轮换刷新令牌
	nowFunc            func() time.Time // 控制 jwt 的时间
}

// NewManagement 定义一个 Management.
// allowTokenHeader: 默认使用 authorization 为认证请求头.
// exposeAccessHeader: 默认使用 x-access-token 为暴露外部的资源请求头.
// exposeRefreshHeader: 默认使用 x-refresh-token 为暴露外部的刷新请求头.
// refreshJWTOptions: 默认使用 nil 为刷新 token 的配置,
// 如要使用 refresh 相关功能则需要使用 WithRefreshJWTOptions 添加相关配置.
// rotateRefreshToken: 默认不轮换刷新令牌.
// 该配置需要设置 refreshJWTOptions 才有效.
func NewManagement[T any](accessJWTOptions Options,
	opts ...option.Option[Management[T]]) *Management[T] {
	dOpts := defaultManagementOptions[T]()
	dOpts.accessJWTOptions = accessJWTOptions
	option.Apply[Management[T]](&dOpts, opts...)

	return &dOpts
}

func defaultManagementOptions[T any]() Management[T] {
	return Management[T]{
		allowTokenHeader:    "authorization",
		exposeAccessHeader:  "x-access-token",
		exposeRefreshHeader: "x-refresh-token",
		rotateRefreshToken:  false,
		nowFunc:             time.Now,
	}
}

// WithAllowTokenHeader 设置允许 token 的请求头.
func WithAllowTokenHeader[T any](header string) option.Option[Management[T]] {
	return func(m *Management[T]) {
		m.allowTokenHeader = header
	}
}

// WithExposeAccessHeader 设置公开资源令牌的请求头.
func WithExposeAccessHeader[T any](header string) option.Option[Management[T]] {
	return func(m *Management[T]) {
		m.exposeAccessHeader = header
	}
}

// WithExposeRefreshHeader 设置公开刷新令牌的请求头.
func WithExposeRefreshHeader[T any](header string) option.Option[Management[T]] {
	return func(m *Management[T]) {
		m.exposeRefreshHeader = header
	}
}

// WithRefreshJWTOptions 设置刷新令牌相关的配置.
func WithRefreshJWTOptions[T any](refreshOpts Options) option.Option[Management[T]] {
	return func(m *Management[T]) {
		m.refreshJWTOptions = &refreshOpts
	}
}

// WithRotateRefreshToken 设置轮换刷新令牌.
func WithRotateRefreshToken[T any](isRotate bool) option.Option[Management[T]] {
	return func(m *Management[T]) {
		m.rotateRefreshToken = isRotate
	}
}

// WithNowFunc 设置当前时间.
// 一般用于测试固定 jwt.
func WithNowFunc[T any](nowFunc func() time.Time) option.Option[Management[T]] {
	return func(m *Management[T]) {
		m.nowFunc = nowFunc
	}
}

// Refresh 刷新 token 的 gin.HandlerFunc.
func (m *Management[T]) Refresh(ctx *gin.Context) {
	if m.refreshJWTOptions == nil {
		slog.Error("refreshJWTOptions 为 nil, 请使用 WithRefreshJWTOptions 设置 refresh 相关的配置")
		ctx.Status(http.StatusInternalServerError)
		return
	}

	tokenStr := m.extractTokenString(ctx)
	clm, err := m.VerifyRefreshToken(tokenStr,
		jwt.WithTimeFunc(m.nowFunc))
	if err != nil {
		slog.Debug("refresh token verification failed")
		ctx.Status(http.StatusUnauthorized)
		return
	}
	accessToken, err := m.GenerateAccessToken(clm.Data)
	if err != nil {
		slog.Error("failed to generate access token")
		ctx.Status(http.StatusInternalServerError)
		return
	}
	ctx.Header(m.exposeAccessHeader, accessToken)

	// 轮换刷新令牌
	if m.rotateRefreshToken {
		refreshToken, err := m.GenerateRefreshToken(clm.Data)
		if err != nil {
			slog.Error("failed to generate refresh token")
			ctx.Status(http.StatusInternalServerError)
			return
		}
		ctx.Header(m.exposeRefreshHeader, refreshToken)
	}
	ctx.Status(http.StatusNoContent)
}

// MiddlewareBuilder 登录认证的中间件.
func (m *Management[T]) MiddlewareBuilder() *MiddlewareBuilder[T] {
	return newMiddlewareBuilder[T](m)
}

// extractTokenString 提取 token 字符串.
func (m *Management[T]) extractTokenString(ctx *gin.Context) string {
	authCode := ctx.GetHeader(m.allowTokenHeader)
	if authCode == "" {
		return ""
	}
	var b strings.Builder
	b.WriteString(bearerPrefix)
	b.WriteString(" ")
	prefix := b.String()
	if strings.HasPrefix(authCode, prefix) {
		return authCode[len(prefix):]
	}
	return ""
}

// GenerateAccessToken 生成资源 token.
func (m *Management[T]) GenerateAccessToken(data T) (string, error) {
	nowTime := m.nowFunc()
	claims := RegisteredClaims[T]{
		Data: data,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.accessJWTOptions.Issuer,
			ExpiresAt: jwt.NewNumericDate(nowTime.Add(m.accessJWTOptions.Expire)),
			IssuedAt:  jwt.NewNumericDate(nowTime),
			ID:        m.accessJWTOptions.genIDFn(),
		},
	}
	token := jwt.NewWithClaims(m.accessJWTOptions.Method, claims)
	return token.SignedString([]byte(m.accessJWTOptions.EncryptionKey))
}

// VerifyAccessToken 校验资源 token.
func (m *Management[T]) VerifyAccessToken(token string, opts ...jwt.ParserOption) (RegisteredClaims[T], error) {
	t, err := jwt.ParseWithClaims(token, &RegisteredClaims[T]{},
		func(*jwt.Token) (interface{}, error) {
			return []byte(m.accessJWTOptions.DecryptKey), nil
		},
		opts...,
	)
	if err != nil || !t.Valid {
		return RegisteredClaims[T]{}, fmt.Errorf("验证失败: %v", err)
	}
	clm, _ := t.Claims.(*RegisteredClaims[T])
	return *clm, nil
}

// GenerateRefreshToken 生成刷新 token.
// 需要设置 refreshJWTOptions 否则返回 errEmptyRefreshOpts 错误.
func (m *Management[T]) GenerateRefreshToken(data T) (string, error) {
	if m.refreshJWTOptions == nil {
		return "", errEmptyRefreshOpts
	}

	nowTime := m.nowFunc()
	claims := RegisteredClaims[T]{
		Data: data,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.refreshJWTOptions.Issuer,
			ExpiresAt: jwt.NewNumericDate(nowTime.Add(m.refreshJWTOptions.Expire)),
			IssuedAt:  jwt.NewNumericDate(nowTime),
			ID:        m.refreshJWTOptions.genIDFn(),
		},
	}

	token := jwt.NewWithClaims(m.refreshJWTOptions.Method, claims)
	return token.SignedString([]byte(m.refreshJWTOptions.EncryptionKey))
}

// VerifyRefreshToken 校验刷新 token.
// 需要设置 refreshJWTOptions 否则返回 errEmptyRefreshOpts 错误.
func (m *Management[T]) VerifyRefreshToken(token string, opts ...jwt.ParserOption) (RegisteredClaims[T], error) {
	if m.refreshJWTOptions == nil {
		return RegisteredClaims[T]{}, errEmptyRefreshOpts
	}
	t, err := jwt.ParseWithClaims(token, &RegisteredClaims[T]{},
		func(*jwt.Token) (interface{}, error) {
			return []byte(m.refreshJWTOptions.DecryptKey), nil
		},
		opts...,
	)
	if err != nil || !t.Valid {
		return RegisteredClaims[T]{}, fmt.Errorf("验证失败: %v", err)
	}
	clm, _ := t.Claims.(*RegisteredClaims[T])
	return *clm, nil
}

// SetClaims 设置 claims 到 key=`claims` 的 gin.Context 中.
func (m *Management[T]) SetClaims(ctx *gin.Context, claims RegisteredClaims[T]) {
	ctx.Set("claims", claims)
}
