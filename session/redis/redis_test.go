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

package redis

import (
	"net/http/httptest"
	"testing"

	"github.com/ecodeclub/ginx/gctx"
	"github.com/ecodeclub/ginx/internal/mocks"
	"github.com/ecodeclub/ginx/session"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestSessionProvider_NewSession(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) redis.Cmdable

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
			}, 123,
				map[string]string{"hello": "world"}, map[string]any{})
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
