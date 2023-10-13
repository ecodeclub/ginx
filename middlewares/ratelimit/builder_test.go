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

package ratelimit

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/ecodeclub/ginx/internal/ratelimit"
	limitmocks "github.com/ecodeclub/ginx/internal/ratelimit/mocks"
)

func TestBuilder_SetKeyGenFunc(t *testing.T) {
	tests := []struct {
		name       string
		reqBuilder func(t *testing.T) *http.Request
		fn         func(*gin.Context) string
		want       string
	}{
		{
			name: "设置key成功",
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.RemoteAddr = "127.0.0.1:80"
				return req
			},
			fn: func(ctx *gin.Context) string {
				return "test"
			},
			want: "test",
		},
		{
			name: "默认key",
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.RemoteAddr = "127.0.0.1:80"
				return req
			},
			want: "ip-limiter:127.0.0.1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBuilder(nil)
			if tt.fn != nil {
				b.SetKeyGenFunc(tt.fn)
			}

			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			req := tt.reqBuilder(t)
			ctx.Request = req

			assert.Equal(t, tt.want, b.genKeyFn(ctx))
		})
	}
}

func TestBuilder_Build(t *testing.T) {
	const limitURL = "/limit"
	tests := []struct {
		name string

		mock       func(ctrl *gomock.Controller) ratelimit.Limiter
		reqBuilder func(t *testing.T) *http.Request

		// 预期响应
		wantCode int
	}{
		{
			name: "不限流",
			mock: func(ctrl *gomock.Controller) ratelimit.Limiter {
				limiter := limitmocks.NewMockLimiter(ctrl)
				limiter.EXPECT().Limit(gomock.Any(), gomock.Any()).
					Return(false, nil)
				return limiter
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, limitURL, nil)
				if err != nil {
					t.Fatal(err)
				}
				return req
			},
			wantCode: http.StatusOK,
		},
		{
			name: "限流",
			mock: func(ctrl *gomock.Controller) ratelimit.Limiter {
				limiter := limitmocks.NewMockLimiter(ctrl)
				limiter.EXPECT().Limit(gomock.Any(), gomock.Any()).
					Return(true, nil)
				return limiter
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, limitURL, nil)
				if err != nil {
					t.Fatal(err)
				}
				return req
			},
			wantCode: http.StatusTooManyRequests,
		},
		{
			name: "系统错误",
			mock: func(ctrl *gomock.Controller) ratelimit.Limiter {
				limiter := limitmocks.NewMockLimiter(ctrl)
				limiter.EXPECT().Limit(gomock.Any(), gomock.Any()).
					Return(false, errors.New("模拟系统错误"))
				return limiter
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, limitURL, nil)
				if err != nil {
					t.Fatal(err)
				}
				return req
			},
			wantCode: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			svc := NewBuilder(tt.mock(ctrl))

			server := gin.Default()
			server.Use(svc.Build())
			svc.RegisterRoutes(server)

			req := tt.reqBuilder(t)
			recorder := httptest.NewRecorder()

			server.ServeHTTP(recorder, req)

			assert.Equal(t, tt.wantCode, recorder.Code)
		})
	}
}

func TestBuilder_limit(t *testing.T) {
	tests := []struct {
		name string

		mock       func(ctrl *gomock.Controller) ratelimit.Limiter
		reqBuilder func(t *testing.T) *http.Request

		// 预期响应
		want    bool
		wantErr error
	}{
		{
			name: "不限流",
			mock: func(ctrl *gomock.Controller) ratelimit.Limiter {
				limiter := limitmocks.NewMockLimiter(ctrl)
				limiter.EXPECT().Limit(gomock.Any(), gomock.Any()).
					Return(false, nil)
				return limiter
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.RemoteAddr = "127.0.0.1:80"
				return req
			},
			want: false,
		},
		{
			name: "限流",
			mock: func(ctrl *gomock.Controller) ratelimit.Limiter {
				limiter := limitmocks.NewMockLimiter(ctrl)
				limiter.EXPECT().Limit(gomock.Any(), gomock.Any()).
					Return(true, nil)
				return limiter
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.RemoteAddr = "127.0.0.1:80"
				return req
			},
			want: true,
		},
		{
			name: "限流代码出错",
			mock: func(ctrl *gomock.Controller) ratelimit.Limiter {
				limiter := limitmocks.NewMockLimiter(ctrl)
				limiter.EXPECT().Limit(gomock.Any(), gomock.Any()).
					Return(false, errors.New("模拟系统错误"))
				return limiter
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.RemoteAddr = "127.0.0.1:80"
				return req
			},
			want:    false,
			wantErr: errors.New("模拟系统错误"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			limiter := tt.mock(ctrl)
			b := NewBuilder(limiter)

			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			req := tt.reqBuilder(t)
			ctx.Request = req

			got, err := b.limit(ctx)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func (b *Builder) RegisterRoutes(server *gin.Engine) {
	server.GET("/limit", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})
}
