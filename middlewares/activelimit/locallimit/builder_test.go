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

package locallimit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalActiveLimit_Build(t *testing.T) {
	const (
		url = "/"
	)
	tests := []struct {
		name        string
		countActive int64
		reqBuilder  func(t *testing.T) *http.Request

		// 预期响应
		wantCode int
	}{
		{
			name:        "正常通过",
			countActive: 0,
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, url, nil)
				require.NoError(t, err)
				return req
			},
			wantCode: http.StatusNoContent,
		},
		{
			name:        "限流中",
			countActive: 1,
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, url, nil)
				require.NoError(t, err)
				return req
			},
			wantCode: http.StatusTooManyRequests,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limit := NewLocalActiveLimit(1)
			limit.countActive.Store(tt.countActive)
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
