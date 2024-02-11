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

package session

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ecodeclub/ginx/gctx"
	"github.com/ecodeclub/ginx/internal/errs"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	p := NewMockProvider(ctrl)
	// 包变量的垃圾之处
	SetDefaultProvider(p)
	defer SetDefaultProvider(nil)
	p.EXPECT().NewSession(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx *gctx.Context, uid int64, jwtData map[string]string,
			sessData map[string]any) (Session, error) {
			return &MemorySession{data: sessData,
				claims: Claims{Uid: uid, Data: jwtData}}, nil
		})
	sess, err := NewSession(new(gctx.Context), 123,
		map[string]string{"jwt": "true"},
		map[string]any{"session": "true"})
	require.NoError(t, err)
	assert.Equal(t, &MemorySession{
		data: map[string]any{"session": "true"},
		claims: Claims{
			Uid:  123,
			Data: map[string]string{"jwt": "true"},
		},
	}, sess)
}

func TestCheckLoginMiddleware(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	p := NewMockProvider(ctrl)
	// 包变量的垃圾之处
	SetDefaultProvider(p)
	defer SetDefaultProvider(nil)
	server := gin.Default()
	server.Use(CheckLoginMiddleware())
	server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "OK")
	})

	// 第一个请求，被拒绝
	p.EXPECT().Get(gomock.Any()).Return(nil, errs.ErrUnauthorized)
	recorder := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "http://localhost/hello", nil)
	require.NoError(t, err)
	server.ServeHTTP(recorder, req)
	assert.Equal(t, 401, recorder.Code)

	// 第二个请求，被处理了

	p.EXPECT().Get(gomock.Any()).Return(NewMemorySession(Claims{}), nil)
	recorder = httptest.NewRecorder()
	req, err = http.NewRequest(http.MethodGet, "http://localhost/hello", nil)
	require.NoError(t, err)
	server.ServeHTTP(recorder, req)
	assert.Equal(t, 200, recorder.Code)
}
