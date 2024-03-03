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

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"

	limit "github.com/ecodeclub/ginx/middlewares/ratelimit"
)

func TestBuilder_e2e_RateLimit(t *testing.T) {
	const (
		ip       = "127.0.0.1"
		limitURL = "/limit"
	)
	rdb := initRedis()
	server := initWebServer(rdb)
	RegisterRoutes(server)

	tests := []struct {
		// 名字
		name string
		// 要提前准备数据
		before func(t *testing.T)
		// 验证并且删除数据
		after func(t *testing.T)

		// 预期响应
		wantCode int
	}{
		{
			name:   "不限流",
			before: func(t *testing.T) {},
			after: func(t *testing.T) {
				rdb.Del(context.Background(), fmt.Sprintf("ip-limiter:%s", ip))
			},
			wantCode: http.StatusOK,
		},
		{
			name: "限流",
			before: func(t *testing.T) {
				req, err := http.NewRequest(http.MethodGet, limitURL, nil)
				req.RemoteAddr = ip + ":80"
				assert.NoError(t, err)
				recorder := httptest.NewRecorder()
				server.ServeHTTP(recorder, req)
			},
			after: func(t *testing.T) {
				rdb.Del(context.Background(), fmt.Sprintf("ip-limiter:%s", ip))
			},
			wantCode: http.StatusTooManyRequests,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer tt.after(t)
			tt.before(t)
			req, err := http.NewRequest(http.MethodGet, limitURL, nil)
			req.RemoteAddr = ip + ":80"
			assert.NoError(t, err)

			recorder := httptest.NewRecorder()
			server.ServeHTTP(recorder, req)

			code := recorder.Code
			assert.Equal(t, tt.wantCode, code)
		})
	}
}

func RegisterRoutes(server *gin.Engine) {
	server.GET("/limit", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})
}

func initRedis() redis.Cmdable {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:16379",
	})
	return redisClient
}

func initWebServer(cmd redis.Cmdable) *gin.Engine {
	server := gin.Default()
	limiter := limit.NewRedisSlidingWindowLimiter(cmd, 500*time.Millisecond, 1)
	server.Use(limit.NewBuilder(limiter).Build())
	return server
}
