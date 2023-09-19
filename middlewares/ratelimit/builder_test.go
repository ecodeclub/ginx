package ratelimit

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/ecodeclub/ginx/ratelimit"
	limitmocks "github.com/ecodeclub/ginx/ratelimit/mocks"
)

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
			// name: "不限流",
			name: "no_limit",
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
			// name: "限流",
			name: "limit",
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
			// name: "系统错误",
			name: "system_error",
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

			// 注册路由
			server := gin.Default()
			server.Use(svc.Build())
			svc.RegisterRoutes(server)

			// 准备请求
			req := tt.reqBuilder(t)
			// 准备记录响应
			recorder := httptest.NewRecorder()
			// 执行
			server.ServeHTTP(recorder, req)
			// 断言
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
				req.Header.Set("X-Real-IP", "127.0.0.1")
				if err != nil {
					t.Fatal(err)
				}
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
				req.Header.Set("X-Real-IP", "127.0.0.1")
				if err != nil {
					t.Fatal(err)
				}
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
