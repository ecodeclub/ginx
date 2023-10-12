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
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ecodeclub/ekit/bean/option"
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

func TestManagement_Refresh(t *testing.T) {
	type testCase[T any] struct {
		name             string
		m                *Management[T]
		reqBuilder       func(t *testing.T) *http.Request
		wantCode         int
		wantAccessToken  string
		wantRefreshToken string
	}
	tests := []testCase[data]{
		{
			// 更新资源令牌并轮换刷新令牌
			name: "refresh_access_token_and_rotate_refresh_token",
			m: NewManagement[data](defaultOption,
				WithRefreshJWTOptions[data](
					NewOptions(24*60*time.Minute,
						"refresh sign key",
					)),
				WithRotateRefreshToken[data](true),
				WithNowFunc[data](func() time.Time {
					return time.UnixMilli(1695623000000)
				}),
			),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/refresh", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add("authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NjU3NjAwLCJpYXQiOjE2OTU1NzEyMDB9.y2AQ98i0le5AbmJFgYCAfCVAphd_9NecmHdhtehMSZE")
				return req
			},
			wantCode:         http.StatusNoContent,
			wantAccessToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NjIzNjAwLCJpYXQiOjE2OTU2MjMwMDB9.i4kCx4-s5EM0a8w2o0usSfkMTLmzUSuEe-inlzg6ru0",
			wantRefreshToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NzA5NDAwLCJpYXQiOjE2OTU2MjMwMDB9.IzPgEwXgoAwaFK-eby4uMl0GYBQwdfZYRi2Bhk3iE_8",
		},
		{
			// 更新资源令牌但轮换刷新令牌生成失败
			name: "refresh_access_token_but_gen_rotate_refresh_token_failed",
			m: NewManagement[data](defaultOption,
				WithRefreshJWTOptions[data](
					NewOptions(24*60*time.Minute,
						"refresh sign key",
						WithMethod(jwt.SigningMethodRS256),
					)),
				WithRotateRefreshToken[data](true),
				WithNowFunc[data](func() time.Time {
					return time.UnixMilli(1695623000000)
				}),
			),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/refresh", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add("authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NjU3NjAwLCJpYXQiOjE2OTU1NzEyMDB9.y2AQ98i0le5AbmJFgYCAfCVAphd_9NecmHdhtehMSZE")
				return req
			},
			wantCode: http.StatusInternalServerError,
		},
		{
			// 更新资源令牌
			name: "refresh_access_token",
			m: NewManagement[data](defaultOption,
				WithRefreshJWTOptions[data](
					NewOptions(24*60*time.Minute,
						"refresh sign key",
					)),
				WithNowFunc[data](func() time.Time {
					return time.UnixMilli(1695623000000)
				}),
			),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/refresh", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add("authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NjU3NjAwLCJpYXQiOjE2OTU1NzEyMDB9.y2AQ98i0le5AbmJFgYCAfCVAphd_9NecmHdhtehMSZE")
				return req
			},
			wantCode:        http.StatusNoContent,
			wantAccessToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NjIzNjAwLCJpYXQiOjE2OTU2MjMwMDB9.i4kCx4-s5EM0a8w2o0usSfkMTLmzUSuEe-inlzg6ru0",
		},
		{
			// 生成资源令牌失败
			name: "gen_access_token_failed",
			m: NewManagement[data](
				Options{
					Expire:        10 * time.Minute,
					EncryptionKey: encryptionKey,
					DecryptKey:    encryptionKey,
					Method:        jwt.SigningMethodRS256,
					genIDFn:       func() string { return "" },
				},
				WithRefreshJWTOptions[data](
					NewOptions(24*60*time.Minute,
						"refresh sign key",
					)),
				WithNowFunc[data](func() time.Time {
					return time.UnixMilli(1695623000000)
				}),
			),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/refresh", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add("authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NjU3NjAwLCJpYXQiOjE2OTU1NzEyMDB9.y2AQ98i0le5AbmJFgYCAfCVAphd_9NecmHdhtehMSZE")
				return req
			},
			wantCode: http.StatusInternalServerError,
		},
		{
			// 刷新令牌认证失败
			name: "refresh_token_verify_failed",
			m: NewManagement[data](
				defaultOption,
				WithRefreshJWTOptions[data](
					NewOptions(24*60*time.Minute,
						"refresh sign key",
					)),
				WithNowFunc[data](func() time.Time {
					return time.UnixMilli(1695723000000)
				}),
			),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/refresh", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add("authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NjU3NjAwLCJpYXQiOjE2OTU1NzEyMDB9.y2AQ98i0le5AbmJFgYCAfCVAphd_9NecmHdhtehMSZE")
				return req
			},
			wantCode: http.StatusUnauthorized,
		},
		{
			// 没有设置刷新令牌选项
			name: "not_set_refreshJWTOptions",
			m: NewManagement[data](
				defaultOption,
				WithNowFunc[data](func() time.Time {
					return time.UnixMilli(1695723000000)
				}),
			),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/refresh", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add("authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NjU3NjAwLCJpYXQiOjE2OTU1NzEyMDB9.y2AQ98i0le5AbmJFgYCAfCVAphd_9NecmHdhtehMSZE")
				return req
			},
			wantCode: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := gin.Default()
			tt.m.registerRoutes(server)

			req := tt.reqBuilder(t)
			recorder := httptest.NewRecorder()

			server.ServeHTTP(recorder, req)
			assert.Equal(t, tt.wantCode, recorder.Code)
			if tt.wantCode != http.StatusOK {
				return
			}
			assert.Equal(t, tt.wantAccessToken,
				recorder.Header().Get("x-access-token"))
			assert.Equal(t, tt.wantRefreshToken,
				recorder.Header().Get("x-refresh-token"))
		})
	}
}

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

func TestManagement_GenerateRefreshToken(t *testing.T) {
	m := defaultManagement
	type testCase[T any] struct {
		name              string
		refreshJWTOptions *Options
		data              T
		want              string
		wantErr           error
	}
	tests := []testCase[data]{
		{
			name: "normal",
			refreshJWTOptions: func() *Options {
				opt := NewOptions(24*60*time.Minute, "refresh sign key")
				return &opt
			}(),
			data: data{Foo: "1"},
			want: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NjU3NjAwLCJpYXQiOjE2OTU1NzEyMDB9.y2AQ98i0le5AbmJFgYCAfCVAphd_9NecmHdhtehMSZE",
		},
		{
			name:    "mistake",
			data:    data{Foo: "1"},
			want:    "",
			wantErr: errEmptyRefreshOpts,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.refreshJWTOptions = tt.refreshJWTOptions
			got, err := m.GenerateRefreshToken(tt.data)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestManagement_VerifyRefreshToken(t *testing.T) {
	defaultRefOpts := Options{
		Expire:        24 * 60 * time.Minute,
		EncryptionKey: "refresh sign key",
		DecryptKey:    "refresh sign key",
		Method:        jwt.SigningMethodHS256,
	}
	type testCase[T any] struct {
		name    string
		m       *Management[T]
		token   string
		want    RegisteredClaims[T]
		wantErr error
	}
	tests := []testCase[data]{
		{
			name: "normal",
			m: NewManagement[data](defaultOption,
				WithNowFunc[data](func() time.Time {
					return time.UnixMilli(1695601200000)
				}),
				WithRefreshJWTOptions[data](defaultRefOpts),
			),
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NjU3NjAwLCJpYXQiOjE2OTU1NzEyMDB9.y2AQ98i0le5AbmJFgYCAfCVAphd_9NecmHdhtehMSZE",
			want: RegisteredClaims[data]{
				Data: data{Foo: "1"},
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(nowTime.Add(24 * 60 * time.Minute)),
					IssuedAt:  jwt.NewNumericDate(nowTime),
				},
			},
		},
		{
			// token 过期了
			name: "token_expired",
			m: NewManagement[data](defaultOption,
				WithNowFunc[data](func() time.Time {
					return time.UnixMilli(1695701200000)
				}),
				WithRefreshJWTOptions[data](defaultRefOpts),
			),
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NjU3NjAwLCJpYXQiOjE2OTU1NzEyMDB9.y2AQ98i0le5AbmJFgYCAfCVAphd_9NecmHdhtehMSZE",
			wantErr: fmt.Errorf("验证失败: %v",
				fmt.Errorf("%v: %v", jwt.ErrTokenInvalidClaims, jwt.ErrTokenExpired)),
		},
		{
			// token 签名错误
			name: "bad_sign_key",
			m: NewManagement[data](defaultOption,
				WithNowFunc[data](func() time.Time {
					return time.UnixMilli(1695601200000)
				}),
				WithRefreshJWTOptions[data](defaultRefOpts),
			),
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NjU3NjAwLCJpYXQiOjE2OTU1NzEyMDB9.yZ_ZlD1jE-0b3qd0bicTDLSdwGsenv6tRmOEqMCM2uw",
			wantErr: fmt.Errorf("验证失败: %v",
				fmt.Errorf("%v: %v", jwt.ErrTokenSignatureInvalid, jwt.ErrSignatureInvalid)),
		},
		{
			// 错误的 token
			name: "bad_token",
			m: NewManagement[data](defaultOption,
				WithNowFunc[data](func() time.Time {
					return time.UnixMilli(1695601200000)
				}),
				WithRefreshJWTOptions[data](defaultRefOpts),
			),
			token: "bad_token",
			wantErr: fmt.Errorf("验证失败: %v: token contains an invalid number of segments",
				jwt.ErrTokenMalformed),
		},
		{
			name: "no_refresh_options",
			m: NewManagement[data](defaultOption,
				WithNowFunc[data](func() time.Time {
					return time.UnixMilli(1695601200000)
				}),
			),
			token:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NjU3NjAwLCJpYXQiOjE2OTU1NzEyMDB9.y2AQ98i0le5AbmJFgYCAfCVAphd_9NecmHdhtehMSZE",
			wantErr: errEmptyRefreshOpts,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.m.VerifyRefreshToken(tt.token,
				jwt.WithTimeFunc(tt.m.nowFunc))
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestManagement_SetClaims(t *testing.T) {
	m := defaultManagement
	type testCase[T any] struct {
		name    string
		claims  RegisteredClaims[T]
		want    RegisteredClaims[T]
		wantErr error
	}
	tests := []testCase[data]{
		{
			name:    "normal",
			claims:  defaultClaims,
			want:    defaultClaims,
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
			m.SetClaims(ctx, tt.claims)
			v, ok := ctx.Get("claims")
			if !ok {
				t.Errorf("claims not found")
			}
			clm, ok := v.(RegisteredClaims[data])
			if !ok {
				t.Errorf("claims type error")
			}
			assert.Equal(t, tt.want, clm)
		})
	}
}

func TestManagement_extractTokenString(t *testing.T) {
	m := defaultManagement
	type header struct {
		key   string
		value string
	}
	type testCase[T any] struct {
		name   string
		header header
		want   string
	}
	tests := []testCase[data]{
		{
			name: "normal",
			header: header{
				key:   "authorization",
				value: "Bearer token",
			},
			want: "token",
		},
		{
			name: "mistake_prefix",
			header: header{
				key:   "authorization",
				value: "bearer token",
			},
		},
		{
			name: "no_allow_token_header",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			req, err := http.NewRequest(http.MethodGet, "", nil)
			req.Header.Add(tt.header.key, tt.header.value)
			if err != nil {
				t.Fatal(err)
			}
			ctx.Request = req

			got := m.extractTokenString(ctx)
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

func TestWithAllowTokenHeader(t *testing.T) {
	type testCase[T any] struct {
		name string
		fn   func() option.Option[Management[T]]
		want string
	}
	tests := []testCase[data]{
		{
			name: "default",
			fn: func() option.Option[Management[data]] {
				return nil
			},
			want: "authorization",
		},
		{
			name: "set_another_header",
			fn: func() option.Option[Management[data]] {
				return WithAllowTokenHeader[data]("jwt")
			},
			want: "jwt",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			if tt.fn() == nil {
				got = NewManagement[data](
					defaultOption,
				).allowTokenHeader
			} else {
				got = NewManagement[data](
					defaultOption,
					tt.fn(),
				).allowTokenHeader
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWithExposeAccessHeader(t *testing.T) {
	type testCase[T any] struct {
		name string
		fn   func() option.Option[Management[T]]
		want string
	}
	tests := []testCase[data]{
		{
			name: "default",
			fn: func() option.Option[Management[data]] {
				return nil
			},
			want: "x-access-token",
		},
		{
			name: "set_another_header",
			fn: func() option.Option[Management[data]] {
				return WithExposeAccessHeader[data]("token")
			},
			want: "token",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			if tt.fn() == nil {
				got = NewManagement[data](
					defaultOption,
				).exposeAccessHeader
			} else {
				got = NewManagement[data](
					defaultOption,
					tt.fn(),
				).exposeAccessHeader
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWithExposeRefreshHeader(t *testing.T) {
	type testCase[T any] struct {
		name string
		fn   func() option.Option[Management[T]]
		want string
	}
	tests := []testCase[data]{
		{
			name: "default",
			fn: func() option.Option[Management[data]] {
				return nil
			},
			want: "x-refresh-token",
		},
		{
			name: "set_another_header",
			fn: func() option.Option[Management[data]] {
				return WithExposeRefreshHeader[data]("refresh-token")
			},
			want: "refresh-token",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			if tt.fn() == nil {
				got = NewManagement[data](
					defaultOption,
				).exposeRefreshHeader
			} else {
				got = NewManagement[data](
					defaultOption,
					tt.fn(),
				).exposeRefreshHeader
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWithRotateRefreshToken(t *testing.T) {
	type testCase[T any] struct {
		name string
		fn   func() option.Option[Management[T]]
		want bool
	}
	tests := []testCase[data]{
		{
			name: "default",
			fn: func() option.Option[Management[data]] {
				return nil
			},
			want: false,
		},
		{
			name: "set_another_header",
			fn: func() option.Option[Management[data]] {
				return WithRotateRefreshToken[data](true)
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got bool
			if tt.fn() == nil {
				got = NewManagement[data](
					defaultOption,
				).rotateRefreshToken
			} else {
				got = NewManagement[data](
					defaultOption,
					tt.fn(),
				).rotateRefreshToken
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWithNowFunc(t *testing.T) {
	type testCase[T any] struct {
		name string
		fn   func() option.Option[Management[T]]
		want time.Time
	}
	tests := []testCase[data]{
		{
			name: "default",
			fn: func() option.Option[Management[data]] {
				return nil
			},
			want: time.Now(),
		},
		{
			name: "set_another_now_func",
			fn: func() option.Option[Management[data]] {
				return WithNowFunc[data](func() time.Time {
					return nowTime
				})
			},
			want: nowTime,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got time.Time
			if tt.fn() == nil {
				got = NewManagement[data](
					defaultOption,
				).nowFunc()
			} else {
				got = NewManagement[data](
					defaultOption,
					tt.fn(),
				).nowFunc()
			}
			assert.Equal(t, tt.want.Unix(), got.Unix())
		})
	}
}

func TestWithRefreshJWTOptions(t *testing.T) {
	var genIDFn func() string
	type testCase[T any] struct {
		name string
		fn   func() option.Option[Management[T]]
		want *Options
	}
	tests := []testCase[data]{
		{
			name: "default",
			fn: func() option.Option[Management[data]] {
				return nil
			},
			want: nil,
		},
		{
			name: "set_refresh_jwt_options",
			fn: func() option.Option[Management[data]] {
				return WithRefreshJWTOptions[data](
					NewOptions(
						24*60*time.Minute,
						"refresh sign key",
						WithGenIDFunc(genIDFn),
					),
				)
			},
			want: &Options{
				Expire:        24 * 60 * time.Minute,
				EncryptionKey: "refresh sign key",
				DecryptKey:    "refresh sign key",
				Method:        jwt.SigningMethodHS256,
				genIDFn:       genIDFn,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got *Options
			if tt.fn() == nil {
				got = NewManagement[data](
					defaultOption,
				).refreshJWTOptions
			} else {
				got = NewManagement[data](
					defaultOption,
					tt.fn(),
				).refreshJWTOptions
			}
			assert.Equal(t, tt.want, got)
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
	server.GET("/refresh", m.Refresh)
}
