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

package redislimit

import (
	"context"
	"errors"
	redismocks "github.com/ecodeclub/ginx/internal/mocks"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRedisActiveLimit_Build(t *testing.T) {
	testCases := []struct {
		name             string
		maxCount         int64
		key              string
		mock             func(ctrl *gomock.Controller, key string) redis.Cmdable
		getReq           func() *http.Request
		createMiddleware func(redisClient redis.Cmdable) gin.HandlerFunc
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
			mock: func(ctrl *gomock.Controller, key string) redis.Cmdable {
				redisClient := redismocks.NewMockCmdable(ctrl)
				res1 := redis.NewIntCmd(context.Background())
				res1.SetErr(nil)
				res1.SetVal(int64(1))
				redisClient.EXPECT().Incr(gomock.Any(), key).Return(res1)

				res2 := redis.NewIntCmd(context.Background())
				res2.SetErr(nil)
				res2.SetVal(int64(0))
				redisClient.EXPECT().Decr(gomock.Any(), key).Return(res2)
				return redisClient

			},
			createMiddleware: func(redisClient redis.Cmdable) gin.HandlerFunc {
				return NewRedisActiveLimit(redisClient, 1, "test").Build()
			},
			getReq: func() *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/activelimit", nil)
				require.NoError(t, err)
				return req
			},
			before: func(server *gin.Engine, key string) {

			},
			after: func(string2 string) (int64, error) {
				return 0, nil
			},
			maxCount: 1,
			key:      "test",
			wantCode: http.StatusOK,
		},
		{
			name: "开启限流,RedisLimit正常操作,但是减一操作异常",
			mock: func(ctrl *gomock.Controller, key string) redis.Cmdable {
				redisClient := redismocks.NewMockCmdable(ctrl)
				res1 := redis.NewIntCmd(context.Background())
				res1.SetErr(nil)
				res1.SetVal(int64(1))
				redisClient.EXPECT().Incr(gomock.Any(), key).Return(res1)

				res2 := redis.NewIntCmd(context.Background())
				res2.SetErr(errors.New("减一操作异常"))
				res2.SetVal(int64(0))
				redisClient.EXPECT().Decr(gomock.Any(), key).Return(res2)
				return redisClient

			},
			createMiddleware: func(redisClient redis.Cmdable) gin.HandlerFunc {
				return NewRedisActiveLimit(redisClient, 1, "test").Build()
			},
			getReq: func() *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/activelimit", nil)
				require.NoError(t, err)
				return req
			},
			before: func(server *gin.Engine, key string) {

			},
			after: func(string2 string) (int64, error) {
				return 0, nil
			},
			maxCount: 1,
			key:      "test",
			wantCode: http.StatusOK,
		},
		{
			name: "开启限流,RedisLimit,有一个人长时间没退出,导致限流",
			mock: func(ctrl *gomock.Controller, key string) redis.Cmdable {
				//第一个进来的
				redisClient := redismocks.NewMockCmdable(ctrl)
				res1 := redis.NewIntCmd(context.Background())
				res1.SetErr(nil)
				res1.SetVal(int64(1))
				redisClient.EXPECT().Incr(gomock.Any(), key).Return(res1)

				//第二个进来的
				res := redis.NewIntCmd(context.Background())
				res.SetErr(nil)
				res.SetVal(int64(2))
				redisClient.EXPECT().Incr(gomock.Any(), key).Return(res)

				//第二个人出去,还有一个处理中
				res2 := redis.NewIntCmd(context.Background())
				res2.SetErr(nil)
				res2.SetVal(int64(1))
				redisClient.EXPECT().Decr(gomock.Any(), key).Return(res2).AnyTimes()

				//res3 := redis.NewIntCmd(context.Background())
				//res3.SetErr(nil)
				//res3.SetVal(int64(0))
				//redisClient.EXPECT().Decr(gomock.Any(), key).Return(res3)

				return redisClient

			},

			createMiddleware: func(redisClient redis.Cmdable) gin.HandlerFunc {
				return NewRedisActiveLimit(redisClient, 1, "test").Build()
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

			interval: time.Millisecond * 20,
			after: func(key string) (int64, error) {
				return 0, nil
			},
			maxCount: 1,
			key:      "test",
			wantCode: http.StatusTooManyRequests,
		},
		{
			name: "开启限流,RedisLimit,有一个人长时间没退出,等待前面退出后,正常请求....",
			mock: func(ctrl *gomock.Controller, key string) redis.Cmdable {
				//第一个进来的
				redisClient := redismocks.NewMockCmdable(ctrl)
				res1 := redis.NewIntCmd(context.Background())
				res1.SetErr(nil)
				res1.SetVal(int64(1))
				redisClient.EXPECT().Incr(gomock.Any(), key).Times(2).Return(res1)
				//第一个人出去
				res2 := redis.NewIntCmd(context.Background())
				res2.SetErr(nil)
				res2.SetVal(int64(0))
				redisClient.EXPECT().Decr(gomock.Any(), key).Times(2).Return(res2)

				return redisClient
			},
			createMiddleware: func(redisClient redis.Cmdable) gin.HandlerFunc {
				return NewRedisActiveLimit(redisClient, 1, "test").Build()
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
			interval: time.Millisecond * 300,
			maxCount: 1,
			key:      "test",
			wantCode: http.StatusOK,
		},
		{
			name: "系统异常",
			mock: func(ctrl *gomock.Controller, key string) redis.Cmdable {
				//第一个进来的
				redisClient := redismocks.NewMockCmdable(ctrl)
				res1 := redis.NewIntCmd(context.Background())
				res1.SetErr(errors.New("redis 异常"))
				res1.SetVal(int64(0))
				redisClient.EXPECT().Incr(gomock.Any(), key).Return(res1)

				return redisClient

			},

			createMiddleware: func(redisClient redis.Cmdable) gin.HandlerFunc {
				return NewRedisActiveLimit(redisClient, 1, "test").Build()
			},
			getReq: func() *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/activelimit", nil)
				require.NoError(t, err)
				return req
			},
			before: func(server *gin.Engine, key string) {

			},

			interval: time.Millisecond * 20,
			after: func(key string) (int64, error) {
				return 0, nil
			},
			maxCount: 1,
			key:      "test",
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctl := gomock.NewController(t)
			defer ctl.Finish()
			server := gin.Default()
			server.Use(tc.createMiddleware(tc.mock(ctl, tc.key)))
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

		})
	}
}
