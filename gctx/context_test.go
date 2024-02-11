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

package gctx

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContext_Query(t *testing.T) {
	testCases := []struct {
		name    string
		req     func(t *testing.T) *http.Request
		key     string
		wantErr error
		wantVal any
	}{
		{
			name: "获得数据",
			req: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "http://localhost/abc?name=123&age=18", nil)
				require.NoError(t, err)
				return req
			},
			key:     "name",
			wantVal: "123",
		},
		{
			name: "没有数据",
			req: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "http://localhost/abc?name=123&age=18", nil)
				require.NoError(t, err)
				return req
			},
			key:     "nickname",
			wantVal: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := &Context{Context: &gin.Context{
				Request: tc.req(t),
			}}
			name := ctx.Query(tc.key)
			val, err := name.String()
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantVal, val)
		})
	}
}

func TestContext_Param(t *testing.T) {
	testCases := []struct {
		name    string
		req     func(t *testing.T) *http.Request
		key     string
		wantErr error
		wantVal any
	}{
		{
			name: "获得数据",
			req: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "http://localhost/hello?name=123&age=18", nil)
				req.Form = url.Values{}
				req.Form.Set("name", "world")
				require.NoError(t, err)
				return req
			},
			key:     "name",
			wantVal: "world",
		},
		{
			name: "没有数据",
			req: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost, "http://localhost/hello?name=123&age=18", nil)
				require.NoError(t, err)
				return req
			},
			key:     "nickname",
			wantVal: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := gin.Default()
			server.POST("/hello", func(context *gin.Context) {
				ctx := &Context{Context: context}
				name := ctx.Param(tc.key)
				val, err := name.String()
				assert.Equal(t, tc.wantErr, err)
				assert.Equal(t, tc.wantVal, val)
			})
			recorder := httptest.NewRecorder()
			server.ServeHTTP(recorder, tc.req(t))
		})
	}
}

func TestContext_Cookie(t *testing.T) {
	testCases := []struct {
		name    string
		req     func(t *testing.T) *http.Request
		key     string
		wantErr error
		wantVal any
	}{
		{
			name: "有cookie",
			req: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "http://localhost/hello?name=123&age=18", nil)
				req.AddCookie(&http.Cookie{
					Name:  "name",
					Value: "world",
				})
				require.NoError(t, err)
				return req
			},
			key:     "name",
			wantVal: "world",
		},
		{
			name: "没有 cookie",
			req: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost, "http://localhost/hello?name=123&age=18", nil)
				require.NoError(t, err)
				return req
			},
			key:     "nickname",
			wantVal: "",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := gin.Default()
			server.POST("/hello", func(context *gin.Context) {
				ctx := &Context{Context: context}
				name := ctx.Param(tc.key)
				val, err := name.String()
				assert.Equal(t, tc.wantErr, err)
				assert.Equal(t, tc.wantVal, val)
			})
			recorder := httptest.NewRecorder()
			server.ServeHTTP(recorder, tc.req(t))
		})
	}
}
