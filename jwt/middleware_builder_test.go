package jwt

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMiddlewareBuilder_Build(t *testing.T) {
	type testCase[T any] struct {
		name       string
		m          *Management[T]
		reqBuilder func(t *testing.T) *http.Request
		wantCode   int
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
			wantCode: http.StatusOK,
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := gin.Default()
			server.Use(tt.m.MiddlewareBuilder().
				IgnorePath("/login").Build())
			tt.m.registerRoutes(server)

			req := tt.reqBuilder(t)
			recorder := httptest.NewRecorder()

			server.ServeHTTP(recorder, req)
			assert.Equal(t, tt.wantCode, recorder.Code)
		})
	}
}
