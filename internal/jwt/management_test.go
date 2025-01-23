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
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

type data struct {
	Foo string `json:"foo"`
}

var (
	defaultExpire = 10 * time.Minute
	defaultClaims = RegisteredClaims[data]{
		Data: data{Foo: "1"},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(nowTime.Add(defaultExpire)),
			IssuedAt:  jwt.NewNumericDate(nowTime),
		},
	}
	encryptionKey     = "sign key"
	nowTime           = time.UnixMilli(1695571200000)
	defaultOption     = NewOptions(defaultExpire, encryptionKey)
	defaultManagement = NewManagement[data](defaultOption,
		WithNowFunc[data](func() time.Time {
			return nowTime
		}),
	)
)

func TestManagement_GenerateAccessToken(t *testing.T) {
	m := defaultManagement
	type testCase[T any] struct {
		name    string
		data    T
		want    string
		wantErr error
	}
	tests := []testCase[data]{
		{
			name: "normal",
			data: data{Foo: "1"},
			want: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NTcxODAwLCJpYXQiOjE2OTU1NzEyMDB9.RMpM5YNgxl9OtCy4lt_JRxv6k8s6plCkthnAV-vbXEQ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := m.GenerateAccessToken(tt.data)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestManagement_VerifyAccessToken(t *testing.T) {
	type testCase[T any] struct {
		name    string
		m       *Management[T]
		token   string
		want    RegisteredClaims[T]
		wantErr error
	}
	tests := []testCase[data]{
		{
			name:  "normal",
			m:     defaultManagement,
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NTcxODAwLCJpYXQiOjE2OTU1NzEyMDB9.RMpM5YNgxl9OtCy4lt_JRxv6k8s6plCkthnAV-vbXEQ",
			want:  defaultClaims,
		},
		{
			// token 过期了
			name: "token_expired",
			m: NewManagement[data](defaultOption,
				WithNowFunc[data](func() time.Time {
					return time.UnixMilli(1695671200000)
				}),
			),
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NTcxODAwLCJpYXQiOjE2OTU1NzEyMDB9.RMpM5YNgxl9OtCy4lt_JRxv6k8s6plCkthnAV-vbXEQ",
			wantErr: fmt.Errorf("验证失败: %v",
				fmt.Errorf("%v: %v", jwt.ErrTokenInvalidClaims, jwt.ErrTokenExpired)),
		},
		{
			// token 签名错误
			name: "bad_sign_key",
			m: NewManagement[data](
				defaultOption,
				WithNowFunc[data](func() time.Time {
					return nowTime
				}),
			),
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NTcxODAwLCJpYXQiOjE2OTU1NzEyMDB9.pnP991l48s_j4fkiZnmh48gjgDGult9Or_wLChHvYp0",
			wantErr: fmt.Errorf("验证失败: %v",
				fmt.Errorf("%v: %v", jwt.ErrTokenSignatureInvalid, jwt.ErrSignatureInvalid)),
		},
		{
			// 错误的 token
			name:  "bad_token",
			m:     defaultManagement,
			token: "bad_token",
			wantErr: fmt.Errorf("验证失败: %v: token contains an invalid number of segments",
				jwt.ErrTokenMalformed),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.m.VerifyAccessToken(tt.token,
				jwt.WithTimeFunc(tt.m.nowFunc))
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewManagement(t *testing.T) {
	type testCase[T any] struct {
		name             string
		accessJWTOptions Options
		wantPanic        bool
	}
	tests := []testCase[data]{
		{
			name:             "normal",
			accessJWTOptions: defaultOption,
			wantPanic:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if err := recover(); err != nil {
					if !tt.wantPanic {
						t.Errorf("期望出现 painc ,但没有")
					}
				}
			}()
			NewManagement[data](tt.accessJWTOptions)
		})
	}
}

func (m *Management[T]) registerRoutes(server *gin.Engine) {
	server.GET("/", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})
	server.GET("/login", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})
}
