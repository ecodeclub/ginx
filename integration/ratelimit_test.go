//go:build e2e

package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"

	"github.com/ecodeclub/ginx/ioc"
	limit "github.com/ecodeclub/ginx/middlewares/ratelimit"
	"github.com/ecodeclub/ginx/ratelimit"
)

func TestBuilder_e2e_RateLimit(t *testing.T) {
	const (
		ip       = "127.0.0.1"
		limitURL = "/limit"
	)
	rdb := ioc.InitRedis()
	server := InitWebServer(rdb)
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
				rdb.Del(context.Background(), "ip-limiter")
			},
			wantCode: http.StatusOK,
		},
		{
			name: "限流",
			before: func(t *testing.T) {
				req, err := http.NewRequest(http.MethodGet, limitURL, nil)
				req.Header.Set("X-Real-IP", ip)
				assert.NoError(t, err)
				recorder := httptest.NewRecorder()
				server.ServeHTTP(recorder, req)
			},
			after: func(t *testing.T) {
				rdb.Del(context.Background(), "ip-limiter")
			},
			wantCode: http.StatusTooManyRequests,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer tt.after(t)
			tt.before(t)
			req, err := http.NewRequest(http.MethodGet, limitURL, nil)
			req.Header.Set("X-Real-IP", ip)
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

func InitWebServer(cmd redis.Cmdable) *gin.Engine {
	server := gin.Default()
	limiter := ratelimit.NewRedisSlidingWindowLimiter(cmd, 500*time.Millisecond, 1)
	server.Use(limit.NewBuilder(limiter).Build())
	return server
}
