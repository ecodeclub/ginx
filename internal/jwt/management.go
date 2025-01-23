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
	"fmt"
	"time"

	"github.com/ecodeclub/ekit/bean/option"
	"github.com/golang-jwt/jwt/v5"
)

var _ Manager[int] = &Management[int]{}

type Management[T any] struct {
	accessJWTOptions Options          // 资源 token 选项
	nowFunc          func() time.Time // 控制 jwt 的时间
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
		nowFunc: time.Now,
	}
}

// WithNowFunc 设置当前时间.
// 一般用于测试固定 jwt.
func WithNowFunc[T any](nowFunc func() time.Time) option.Option[Management[T]] {
	return func(m *Management[T]) {
		m.nowFunc = nowFunc
	}
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
