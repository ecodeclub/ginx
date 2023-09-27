//go:build e2e

package integration

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
	"github.com/stretchr/testify/require"

	"github.com/ecodeclub/ginx/internal/activelimit"
	"github.com/ecodeclub/ginx/internal/activelimit/local_limit"
	"github.com/ecodeclub/ginx/internal/activelimit/redis_limit"
	activelimit2 "github.com/ecodeclub/ginx/middlewares/activelimit"
)

func TestBuilder_e2e_ActiveLocalLimit(t *testing.T) {

	testCases := []struct {
		name             string
		maxCount         int64
		key              string
		getReq           func() *http.Request
		createLimit      func() activelimit.Limiter
		createMiddleware func(limiter activelimit.Limiter) gin.HandlerFunc
		before           func(localLimit activelimit.Limiter, key string, maxCount int64)

		after func(localLimit activelimit.Limiter, key string) (int64, error)
		//响应的code
		wantCode int

		//检查退出的时候redis 状态
		afterCount int64
		afterErr   error
	}{
		{
			name: "开启限流,LocalLimit正常操作",

			createLimit: func() activelimit.Limiter {
				return local_limit.NewLocalLimit()
			},
			createMiddleware: func(limiter activelimit.Limiter) gin.HandlerFunc {
				return activelimit2.NewActiveLimit(1, limiter, func(ctx *gin.Context) string {
					return "test"
				}).Build()
			},
			getReq: func() *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/activelimit", nil)
				require.NoError(t, err)
				return req
			},
			before: func(localLimit activelimit.Limiter, key string, maxCount int64) {

			},
			after: func(localLimit activelimit.Limiter, key string) (int64, error) {
				return 0, nil
			},

			maxCount: 1,
			key:      "test",
			wantCode: 200,

			afterCount: 0,
			afterErr:   nil,
		},
		{
			name: "开启限流,LocalLimit 有一个人很久没出来,新请求被限流",

			createLimit: func() activelimit.Limiter {
				return local_limit.NewLocalLimit()
			},
			createMiddleware: func(limiter activelimit.Limiter) gin.HandlerFunc {
				return activelimit2.NewActiveLimit(1, limiter, func(ctx *gin.Context) string {
					return "test"
				}).Build()
			},
			getReq: func() *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/activelimit", nil)
				require.NoError(t, err)
				return req
			},
			before: func(localLimit activelimit.Limiter, key string, maxCount int64) {
				limited, err := localLimit.Add(context.Background(), key, maxCount)
				assert.Equal(t, limited, false)
				assert.Equal(t, err, nil)
			},
			after: func(localLimit activelimit.Limiter, key string) (int64, error) {
				return 0, nil
			},

			maxCount: 1,
			key:      "test",
			wantCode: http.StatusTooManyRequests,

			afterCount: 0,
			afterErr:   nil,
		},

		{
			name: "关闭限流,LocalLimit 有一个人很久没出来,新请求正常返回",

			createLimit: func() activelimit.Limiter {
				return local_limit.NewLocalLimit()
			},
			createMiddleware: func(limiter activelimit.Limiter) gin.HandlerFunc {
				return activelimit2.NewActiveLimit(1, limiter, func(ctx *gin.Context) string {
					return "test"
				}).SetStatue(false).Build()
			},
			getReq: func() *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/activelimit", nil)
				require.NoError(t, err)
				return req
			},
			before: func(localLimit activelimit.Limiter, key string, maxCount int64) {
				limited, err := localLimit.Add(context.Background(), key, maxCount)
				assert.Equal(t, limited, false)
				assert.Equal(t, err, nil)
			},
			after: func(localLimit activelimit.Limiter, key string) (int64, error) {
				return 0, nil
			},

			maxCount: 1,
			key:      "test",
			wantCode: http.StatusOK,

			afterCount: 0,
			afterErr:   nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			server := gin.Default()
			l := tc.createLimit()
			server.Use(tc.createMiddleware(l))
			server.GET("/activelimit", func(ctx *gin.Context) {
				ctx.Status(http.StatusOK)
			})
			resp := httptest.NewRecorder()
			tc.before(l, tc.key, tc.maxCount)
			server.ServeHTTP(resp, tc.getReq())
			assert.Equal(t, tc.wantCode, resp.Code)

			afterCount, err := tc.after(l, tc.key)

			assert.Equal(t, tc.afterCount, afterCount)
			assert.Equal(t, tc.afterErr, err)
		})
	}

}

func TestBuilder_e2e_ActiveRedisLimit(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:16379",
		Password: "",
		DB:       0,
	})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	err := redisClient.Ping(ctx).Err()
	if err != nil {
		panic("redis  连接失败")
	}
	defer func() {
		_ = redisClient.Close()
	}()

	testCases := []struct {
		name             string
		maxCount         int64
		key              string
		getReq           func() *http.Request
		createLimit      func() activelimit.Limiter
		createMiddleware func(limiter activelimit.Limiter) gin.HandlerFunc
		before           func(localLimit activelimit.Limiter, key string, maxCount int64)

		after func(localLimit activelimit.Limiter, key string) (int64, error)

		//响应的code
		wantCode int

		//检查退出的时候redis 状态
		afterCount int64
		afterErr   error
	}{
		{
			name: "RedisLimit正常操作",

			createLimit: func() activelimit.Limiter {
				return redis_limit.NewRedisLimit(redisClient)
			},
			createMiddleware: func(limiter activelimit.Limiter) gin.HandlerFunc {
				return activelimit2.NewActiveLimit(1, limiter, func(ctx *gin.Context) string {
					return "test"
				}).Build()
			},
			getReq: func() *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/activelimit", nil)
				require.NoError(t, err)
				return req
			},
			before: func(localLimit activelimit.Limiter, key string, maxCount int64) {
				//limited, err := localLimit.Add(context.Background(), key, maxCount)
				//assert.Equal(t, limited, false)
				//assert.Equal(t, err, nil)
			},
			after: func(localLimit activelimit.Limiter, key string) (int64, error) {
				defer func() {
					err = redisClient.Del(context.Background(), key).Err()
					assert.Equal(t, err, nil)
				}()
				return redisClient.Get(context.Background(), key).Int64()
			},

			maxCount: 1,
			key:      "test",
			wantCode: http.StatusOK,

			afterCount: 0,
			afterErr:   nil,
		},
		{
			name: "RedisLimit,有一个人长时间没退出,导致限流",

			createLimit: func() activelimit.Limiter {
				return redis_limit.NewRedisLimit(redisClient)
			},
			createMiddleware: func(limiter activelimit.Limiter) gin.HandlerFunc {
				return activelimit2.NewActiveLimit(1, limiter, func(ctx *gin.Context) string {
					return "test"
				}).Build()
			},
			getReq: func() *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/activelimit", nil)
				require.NoError(t, err)
				return req
			},
			before: func(localLimit activelimit.Limiter, key string, maxCount int64) {
				limited, err := localLimit.Add(context.Background(), key, maxCount)
				assert.Equal(t, limited, false)
				assert.Equal(t, err, nil)
				limited, err = localLimit.Add(context.Background(), key, maxCount)
				assert.Equal(t, limited, true)
				assert.Equal(t, err, nil)
			},
			after: func(localLimit activelimit.Limiter, key string) (int64, error) {
				defer func() {
					err = redisClient.Del(context.Background(), key).Err()
					assert.Equal(t, err, nil)
				}()
				return redisClient.Get(context.Background(), key).Int64()
			},
			maxCount:   1,
			key:        "test",
			wantCode:   http.StatusTooManyRequests,
			afterCount: 1,
			afterErr:   nil,
		},
		{
			name: "RedisLimit,没有开启限流,有一个人长时间没退出,不会限流",

			createLimit: func() activelimit.Limiter {
				return redis_limit.NewRedisLimit(redisClient)
			},
			createMiddleware: func(limiter activelimit.Limiter) gin.HandlerFunc {
				return activelimit2.NewActiveLimit(1, limiter, func(ctx *gin.Context) string {
					return "test"
				}).SetStatue(false).Build()
			},
			getReq: func() *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/activelimit", nil)
				require.NoError(t, err)
				return req
			},
			before: func(localLimit activelimit.Limiter, key string, maxCount int64) {
				limited, err := localLimit.Add(context.Background(), key, maxCount)
				assert.Equal(t, limited, false)
				assert.Equal(t, err, nil)
				limited, err = localLimit.Add(context.Background(), key, maxCount)
				assert.Equal(t, limited, true)
				assert.Equal(t, err, nil)
			},
			after: func(localLimit activelimit.Limiter, key string) (int64, error) {
				defer func() {
					err = redisClient.Del(context.Background(), key).Err()
					assert.Equal(t, err, nil)
				}()
				return redisClient.Get(context.Background(), key).Int64()
			},
			maxCount:   1,
			key:        "test",
			wantCode:   http.StatusOK,
			afterCount: 1,
			afterErr:   nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			server := gin.Default()
			l := tc.createLimit()
			server.Use(tc.createMiddleware(l))
			server.GET("/activelimit", func(ctx *gin.Context) {
				ctx.Status(http.StatusOK)
			})
			resp := httptest.NewRecorder()
			tc.before(l, tc.key, tc.maxCount)

			fmt.Println(redisClient.Get(context.Background(), "test").Int64())
			server.ServeHTTP(resp, tc.getReq())
			assert.Equal(t, tc.wantCode, resp.Code)

			afterCount, err := tc.after(l, tc.key)

			assert.Equal(t, tc.afterCount, afterCount)
			assert.Equal(t, tc.afterErr, err)
		})
	}
}
