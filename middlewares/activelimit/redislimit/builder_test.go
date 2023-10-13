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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/ecodeclub/ginx/internal/mocks"
)

func TestRedisActiveLimit_Build(t *testing.T) {
	const (
		url = "/"
		key = "limit"
	)
	tests := []struct {
		name       string
		mock       func(ctrl *gomock.Controller) redis.Cmdable
		reqBuilder func(t *testing.T) *http.Request

		// 预期响应
		wantCode int
	}{
		{
			name: "正常通过",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewIntCmd(context.Background())
				res.SetVal(int64(1))
				cmd.EXPECT().Incr(gomock.Any(), key).Return(res)
				res = redis.NewIntCmd(context.Background())
				cmd.EXPECT().Decr(gomock.Any(), key).Return(res)
				return cmd
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, url, nil)
				require.NoError(t, err)
				return req
			},
			wantCode: http.StatusNoContent,
		},
		{
			name: "限流中",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				ctx := context.Background()
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewIntCmd(ctx)
				res.SetVal(int64(2))
				cmd.EXPECT().Incr(gomock.Any(), key).Return(res)
				res = redis.NewIntCmd(ctx)
				cmd.EXPECT().Decr(gomock.Any(), key).Return(res)
				return cmd
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, url, nil)
				require.NoError(t, err)
				return req
			},
			wantCode: http.StatusTooManyRequests,
		},
		{
			name: "defer 的减1操作失败",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				ctx := context.Background()
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewIntCmd(ctx)
				res.SetVal(int64(1))
				cmd.EXPECT().Incr(gomock.Any(), key).Return(res)
				res = redis.NewIntCmd(ctx)
				res.SetErr(errors.New("模拟 redis 操作失败"))
				cmd.EXPECT().Decr(gomock.Any(), key).Return(res)
				return cmd
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, url, nil)
				require.NoError(t, err)
				return req
			},
			wantCode: http.StatusNoContent,
		},
		{
			name: "刚进入中间件的加1操作失败",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewIntCmd(context.Background())
				res.SetErr(errors.New("模拟 redis 操作失败"))
				cmd.EXPECT().Incr(gomock.Any(), key).Return(res)
				return cmd
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, url, nil)
				require.NoError(t, err)
				return req
			},
			wantCode: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			limit := NewRedisActiveLimit(tt.mock(ctrl), 1, key)
			server := gin.Default()
			server.Use(limit.Build())
			server.GET(url, func(c *gin.Context) {
				c.Status(http.StatusNoContent)
			})

			req := tt.reqBuilder(t)
			recorder := httptest.NewRecorder()

			server.ServeHTTP(recorder, req)

			assert.Equal(t, tt.wantCode, recorder.Code)
		})
	}
}
