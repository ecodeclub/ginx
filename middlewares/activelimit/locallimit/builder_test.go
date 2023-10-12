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

package locallimit

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLocalActiveLimit_Build(t *testing.T) {

	testCases := []struct {
		name             string
		maxCount         int64
		getReq           func() *http.Request
		createMiddleware func(maxActive int64) gin.HandlerFunc
		before           func(server *gin.Engine)

		after func()
		//响应的code
		wantCode int
		//
		interval time.Duration
	}{
		{
			name: "开启限流,LocalLimit正常操作",

			createMiddleware: func(maxActive int64) gin.HandlerFunc {
				return NewLocalActiveLimit(maxActive).Build()
			},
			getReq: func() *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/activelimit", nil)
				require.NoError(t, err)
				return req
			},
			before: func(server *gin.Engine) {

			},
			after: func() {

			},

			maxCount: 1,
			wantCode: 200,
		},
		{
			name: "开启限流,LocalLimit 有一个人很久没出来,新请求被限流",

			createMiddleware: func(maxActive int64) gin.HandlerFunc {
				return NewLocalActiveLimit(maxActive).Build()
			},
			getReq: func() *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/activelimit", nil)
				require.NoError(t, err)
				return req
			},
			before: func(server *gin.Engine) {
				req, err := http.NewRequest(http.MethodGet, "/activelimit3", nil)
				require.NoError(t, err)
				resp := httptest.NewRecorder()
				server.ServeHTTP(resp, req)
				assert.Equal(t, 200, resp.Code)
			},
			after: func() {

			},

			maxCount: 1,
			wantCode: http.StatusTooManyRequests,
		},
		{
			name: "开启限流,LocalLimit 有一个人很久没出来,等待前面的请求退出后,成功通过",

			createMiddleware: func(maxActive int64) gin.HandlerFunc {
				return NewLocalActiveLimit(maxActive).Build()
			},
			getReq: func() *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/activelimit", nil)
				require.NoError(t, err)
				return req
			},
			before: func(server *gin.Engine) {
				req, err := http.NewRequest(http.MethodGet, "/activelimit3", nil)
				require.NoError(t, err)
				resp := httptest.NewRecorder()
				server.ServeHTTP(resp, req)
				assert.Equal(t, 200, resp.Code)
			},
			after: func() {

			},
			interval: time.Millisecond * 600,
			maxCount: 1,
			wantCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			server := gin.Default()
			server.Use(tc.createMiddleware(tc.maxCount))
			server.GET("/activelimit", func(ctx *gin.Context) {
				ctx.Status(http.StatusOK)
			})
			server.GET("/activelimit3", func(ctx *gin.Context) {
				time.Sleep(time.Millisecond * 300)
				ctx.Status(http.StatusOK)
			})
			resp := httptest.NewRecorder()
			go func() {
				tc.before(server)
			}()
			//加延时保证 tc.before 执行
			time.Sleep(time.Millisecond * 10)

			time.Sleep(tc.interval)
			server.ServeHTTP(resp, tc.getReq())
			assert.Equal(t, tc.wantCode, resp.Code)

			tc.after()

		})
	}

}
