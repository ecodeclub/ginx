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

//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ecodeclub/ginx/middlewares/activelimit/redislimit"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuilder_e2e_ActiveRedisLimit(t *testing.T) {
	redisClient := newRedisTestClient()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	err := redisClient.Ping(ctx).Err()
	if err != nil {
		panic("redislimit  连接失败")
	}
	defer func() {
		_ = redisClient.Close()
	}()

	testCases := []struct {
		name             string
		maxCount         int64
		key              string
		getReq           func() *http.Request
		createMiddleware func(maxActive int64, key string) gin.HandlerFunc
		before           func(server *gin.Engine, key string)

		interval time.Duration
		after    func(string2 string) (int64, error)

		//响应的code
		wantCode int

		//检查退出的时候redis 状态
		afterCount int64
		afterErr   error
	}{
		{
			name: "开启限流,RedisLimit正常操作",

			createMiddleware: func(maxActive int64, key string) gin.HandlerFunc {
				return redislimit.NewRedisActiveLimit(redisClient, maxActive, key).Build()
			},
			getReq: func() *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/activelimit", nil)
				require.NoError(t, err)
				return req
			},
			before: func(server *gin.Engine, key string) {

			},
			interval: time.Millisecond * 10,
			after: func(key string) (int64, error) {

				return redisClient.Get(context.Background(), key).Int64()
			},

			maxCount: 1,
			key:      "test",
			wantCode: http.StatusOK,

			afterCount: 0,
			afterErr:   nil,
		},
		{
			name: "开启限流,RedisLimit,有一个人长时间没退出,导致限流",

			createMiddleware: func(maxActive int64, key string) gin.HandlerFunc {
				return redislimit.NewRedisActiveLimit(redisClient, maxActive, key).Build()
			},
			getReq: func() *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/activelimit", nil)
				require.NoError(t, err)
				return req
			},
			before: func(server *gin.Engine, key string) {

				req, err := http.NewRequest(http.MethodGet, "/activelimit3", nil)
				require.NoError(t, err)
				resp := httptest.NewRecorder()
				server.ServeHTTP(resp, req)
				assert.Equal(t, 200, resp.Code)
			},

			interval: time.Millisecond * 50,
			after: func(key string) (int64, error) {

				return redisClient.Get(context.Background(), key).Int64()
			},
			maxCount:   1,
			key:        "test",
			wantCode:   http.StatusTooManyRequests,
			afterCount: 1,
			afterErr:   nil,
		},
		{
			name: "开启限流,RedisLimit,有一个人长时间没退出,等待前面退出后,正常请求....",

			createMiddleware: func(maxActive int64, key string) gin.HandlerFunc {
				return redislimit.NewRedisActiveLimit(redisClient, maxActive, key).Build()
			},
			getReq: func() *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/activelimit", nil)
				require.NoError(t, err)
				return req
			},
			before: func(server *gin.Engine, key string) {
				req, err := http.NewRequest(http.MethodGet, "/activelimit3", nil)
				require.NoError(t, err)
				resp := httptest.NewRecorder()
				server.ServeHTTP(resp, req)
				assert.Equal(t, 200, resp.Code)
			},
			interval: time.Millisecond * 200,
			after: func(key string) (int64, error) {

				return redisClient.Get(context.Background(), key).Int64()
			},
			maxCount:   1,
			key:        "test",
			wantCode:   http.StatusOK,
			afterCount: 0,
			afterErr:   nil,
		},
	}

	for _, tc := range testCases {
		//这里延时的原因是 保证builder 中的defer 延时操作不会导致测试的异常
		time.Sleep(time.Millisecond * 100)
		redisClient.Del(context.Background(), tc.key)
		fmt.Println(redisClient.Get(context.Background(), tc.key).Int64())
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			server := gin.Default()
			server.Use(tc.createMiddleware(tc.maxCount, tc.key))
			server.GET("/activelimit", func(ctx *gin.Context) {
				ctx.Status(http.StatusOK)
			})
			server.GET("/activelimit3", func(ctx *gin.Context) {
				time.Sleep(time.Millisecond * 100)
				ctx.Status(http.StatusOK)
			})
			resp := httptest.NewRecorder()
			go func() {
				tc.before(server, tc.key)
			}()
			time.Sleep(tc.interval)
			server.ServeHTTP(resp, tc.getReq())
			assert.Equal(t, tc.wantCode, resp.Code)

			afterCount, err := tc.after(tc.key)

			assert.Equal(t, tc.afterCount, afterCount)
			assert.Equal(t, tc.afterErr, err)
		})
	}
}
