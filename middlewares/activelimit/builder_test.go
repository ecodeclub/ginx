package activelimit

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	activeLimitmocks "github.com/ecodeclub/ginx/internal/activelimit/mocks"
)

func TestActiveLimit_Build(t *testing.T) {
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) *ActiveLimit
		getReq   func() *http.Request
		wantCode int
	}{
		{
			name: "开启限流,系统异常",
			mock: func(ctrl *gomock.Controller) *ActiveLimit {
				limit := activeLimitmocks.NewMockLimiter(ctrl)
				limit.EXPECT().Add(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					false, errors.New("限流判断异常"))

				al := NewActiveLimit(1, limit, func(ctx *gin.Context) string {
					return "active-req"
				})
				return al
			},
			getReq: func() *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/activelimit", nil)
				require.NoError(t, err)
				return req
			},
			wantCode: 500,
		},
		{
			name: "开启限流,被限流了",
			mock: func(ctrl *gomock.Controller) *ActiveLimit {
				limit := activeLimitmocks.NewMockLimiter(ctrl)
				limit.EXPECT().Add(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					true, nil)
				al := NewActiveLimit(1, limit, func(ctx *gin.Context) string {
					return "active-req"
				})
				return al
			},
			getReq: func() *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/activelimit", nil)
				require.NoError(t, err)
				return req
			},
			wantCode: 429,
		},
		{
			name: "开启限流,正常请求,但是减一操作异常",
			mock: func(ctrl *gomock.Controller) *ActiveLimit {
				limit := activeLimitmocks.NewMockLimiter(ctrl)
				limit.EXPECT().Add(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					false, nil)
				limit.EXPECT().Sub(gomock.Any(), gomock.Any()).Return(
					true, errors.New("减一操作异常"))
				al := NewActiveLimit(1, limit, func(ctx *gin.Context) string {
					return "active-req"
				})
				return al
			},
			getReq: func() *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/activelimit", nil)
				require.NoError(t, err)
				return req
			},
			wantCode: 200,
		},
		{
			name: "开启限流,正常请求,但是减一操作异常2",
			mock: func(ctrl *gomock.Controller) *ActiveLimit {
				limit := activeLimitmocks.NewMockLimiter(ctrl)
				limit.EXPECT().Add(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					false, nil)
				limit.EXPECT().Sub(gomock.Any(), gomock.Any()).Return(
					false, nil)
				al := NewActiveLimit(1, limit, func(ctx *gin.Context) string {
					return "active-req"
				})
				return al
			},
			getReq: func() *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/activelimit", nil)
				require.NoError(t, err)
				return req
			},
			wantCode: 200,
		},
		{
			name: "开启限流,正常请求,全部正常",
			mock: func(ctrl *gomock.Controller) *ActiveLimit {
				limit := activeLimitmocks.NewMockLimiter(ctrl)
				limit.EXPECT().Add(gomock.Any(), gomock.Any(), gomock.Any()).Return(
					false, nil)
				limit.EXPECT().Sub(gomock.Any(), gomock.Any()).Return(
					true, nil)
				al := NewActiveLimit(1, limit, func(ctx *gin.Context) string {
					return "active-req"
				})
				return al
			},
			getReq: func() *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/activelimit", nil)
				require.NoError(t, err)
				return req
			},
			wantCode: 200,
		},
		{
			name: "关闭限流,正常请求",
			mock: func(ctrl *gomock.Controller) *ActiveLimit {
				limit := activeLimitmocks.NewMockLimiter(ctrl)
				al := NewActiveLimit(1, limit, func(ctx *gin.Context) string {
					return "active-req"
				})
				return al.SetStatue(false)
			},
			getReq: func() *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/activelimit", nil)
				require.NoError(t, err)
				return req
			},
			wantCode: 200,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctl := gomock.NewController(t)

			server := gin.Default()

			server.Use(tc.mock(ctl).Build())
			server.GET("/activelimit", func(ctx *gin.Context) {
				ctx.Status(http.StatusOK)
			})
			resp := httptest.NewRecorder()

			server.ServeHTTP(resp, tc.getReq())
			assert.Equal(t, tc.wantCode, resp.Code)

		})
	}
}
