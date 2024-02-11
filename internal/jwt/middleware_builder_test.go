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
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMiddlewareBuilder_Build(t *testing.T) {
	type testCase[T any] struct {
		name        string
		m           *Management[T]
		reqBuilder  func(t *testing.T) *http.Request
		isUseIgnore bool
		wantCode    int
	}
	tests := []testCase[data]{
		{
			// 验证失败
			name: "verify_failed",
			m:    NewManagement[data](defaultOption),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add("authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NTcxODAwLCJpYXQiOjE2OTU1NzEyMDB9.RMpM5YNgxl9OtCy4lt_JRxv6k8s6plCkthnAV-vbXEQ")
				return req
			},
			wantCode: http.StatusUnauthorized,
		},
		{
			// 提取 token 失败
			name: "extract_token_failed",
			m:    NewManagement[data](defaultOption),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add("authorization", "Bearer ")
				return req
			},
			wantCode: http.StatusUnauthorized,
		},
		{
			// 验证通过
			name: "pass_the_verification",
			m: NewManagement[data](defaultOption,
				WithNowFunc[data](func() time.Time {
					return time.UnixMilli(1695571500000)
				}),
			),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add("authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjp7ImZvbyI6IjEifSwiZXhwIjoxNjk1NTcxODAwLCJpYXQiOjE2OTU1NzEyMDB9.RMpM5YNgxl9OtCy4lt_JRxv6k8s6plCkthnAV-vbXEQ")
				return req
			},
			wantCode: http.StatusOK,
		},
		{
			// 无需认证直接通过
			name: "pass_without_authentication",
			m:    NewManagement[data](defaultOption),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/login", nil)
				if err != nil {
					t.Fatal(err)
				}
				return req
			},
			isUseIgnore: true,
			wantCode:    http.StatusOK,
		},
		{
			// 未使用忽略选项则进行拦截
			name: "intercept_if_ignore_opt_is_not_used",
			m:    NewManagement[data](defaultOption),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/login", nil)
				if err != nil {
					t.Fatal(err)
				}
				return req
			},
			wantCode: http.StatusUnauthorized,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := gin.Default()
			m := tt.m.MiddlewareBuilder()
			if tt.isUseIgnore {
				m = m.IgnorePath("/login")
			}
			server.Use(m.Build())
			tt.m.registerRoutes(server)

			req := tt.reqBuilder(t)
			recorder := httptest.NewRecorder()

			server.ServeHTTP(recorder, req)
			assert.Equal(t, tt.wantCode, recorder.Code)
		})
	}
}
