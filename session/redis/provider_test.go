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

package redis

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ecodeclub/ginx/internal/e2e"
	"github.com/stretchr/testify/suite"

	"github.com/ecodeclub/ginx/gctx"
	"github.com/ecodeclub/ginx/internal/mocks"
	"github.com/ecodeclub/ginx/session"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type ProviderTestSuite struct {
	e2e.BaseSuite
}

func (s *ProviderTestSuite) TestRenewSession() {
	sp := NewSessionProvider(s.RDB, "session")
	req, err := http.NewRequest(http.MethodGet, "localhost:8080/hello", nil)
	require.NoError(s.T(), err)
	writer := httptest.NewRecorder()
	gxCtx := &gctx.Context{
		Context: &gin.Context{
			Request: req,
			Writer:  &e2e.GinResponseWriter{ResponseWriter: writer},
		},
	}
	sess, err := sp.NewSession(gxCtx, 123, map[string]string{"jwtKey1": "jwtVal1"}, map[string]any{"sessKe1": "sessVal1"})
	require.NoError(s.T(), err)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err = sess.Set(ctx, "sessKey2", "sessVal2")
	require.NoError(s.T(), err)
	// 先把 refresh token 取出来，放过去 req 的 header，从而模拟 renew 的请求
	rt := writer.Header().Get("X-Refresh-Token")
	req.Header.Set("Authorization", "Bearer "+rt)
	err = sp.RenewAccessToken(gxCtx)
	require.NoError(s.T(), err)
}

func TestSessionProvider_UpdateClaims(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) redis.Cmdable
		wantErr error
	}{
		{
			name: "更新成功",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				pip := mocks.NewMockPipeliner(ctrl)
				pip.EXPECT().HMSet(gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes().Return(nil)
				pip.EXPECT().Expire(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				pip.EXPECT().Exec(gomock.Any()).Return(nil, nil)
				cmd.EXPECT().Pipeline().Return(pip)
				return cmd
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			client := tc.mock(ctrl)
			sp := NewSessionProvider(client, "123")
			recorder := httptest.NewRecorder()

			ctx, _ := gin.CreateTestContext(recorder)
			// 先创建一个
			_, err := sp.NewSession(&gctx.Context{
				Context: ctx,
			}, 123, map[string]string{"hello": "world"}, map[string]any{})

			gtx := &gctx.Context{
				Context: ctx,
			}
			newCl := session.Claims{
				Uid:  234,
				SSID: "ssid_123",
				Data: map[string]string{"hello": "nihao"}}

			err = sp.UpdateClaims(gtx, newCl)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			token := ctx.Writer.Header().Get("X-Access-Token")
			rc, err := sp.m.VerifyAccessToken(token)
			require.NoError(t, err)
			cl := rc.Data
			assert.Equal(t, newCl, cl)
			token = ctx.Writer.Header().Get("X-Refresh-Token")
			rc, err = sp.m.VerifyAccessToken(token)
			require.NoError(t, err)
			cl = rc.Data
			assert.Equal(t, newCl, cl)
		})
	}
}

func TestProvider(t *testing.T) {
	suite.Run(t, new(ProviderTestSuite))
}

// 历史测试，后面考虑删了
func TestSessionProvider_NewSession(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) redis.Cmdable
		key     string
		wantErr error
	}{
		{
			name: "成功",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				pip := mocks.NewMockPipeliner(ctrl)
				pip.EXPECT().HMSet(gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes().Return(nil)
				pip.EXPECT().Expire(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				pip.EXPECT().Exec(gomock.Any()).Return(nil, nil)
				cmd.EXPECT().Pipeline().Return(pip)
				return cmd
			},
			key: "key1",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			client := tc.mock(ctrl)
			sp := NewSessionProvider(client, tc.key)
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			sess, err := sp.NewSession(&gctx.Context{
				Context: ctx,
			}, 123, map[string]string{"hello": "world"}, map[string]any{})
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			rs, ok := sess.(*Session)
			require.True(t, ok)
			cl := rs.Claims()
			assert.True(t, len(cl.SSID) > 0)
			cl.SSID = ""
			assert.Equal(t, session.Claims{
				Uid:  123,
				Data: map[string]string{"hello": "world"},
			}, cl)
		})
	}
}
