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

package crawlerdetect

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func Test_Builder(t *testing.T) {
	testCases := []struct {
		name string

		reqBuilder func(t *testing.T) *http.Request

		wantCode int
	}{
		{
			name: "空 ip",
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/test", nil)
				if err != nil {
					t.Fatal(err)
				}
				return req
			},
			wantCode: 403,
		},
		{
			name: "无效 ip",
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/test", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Set("X-Forwarded-For", "256.0.0.0")
				req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Baiduspider/2.0; +http://www.baidu.com/search/spider.html)")
				return req
			},
			wantCode: 500,
		},
		{
			name: "用户",
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/test", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Set("X-Forwarded-For", "155.206.198.69")
				req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.82 Safari/537.36")
				return req
			},
			wantCode: 403,
		},
		{
			name: "百度 - Baiduspider",
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/test", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Baiduspider/2.0; +http://www.baidu.com/search/spider.html)")
				req.Header.Set("X-Forwarded-For", "111.206.198.69")
				return req
			},
			wantCode: 200,
		},
		{
			name: "百度 - Baiduspider-render",
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/test", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone;CPU iPhone OS 9_1 like Mac OS X) AppleWebKit/601.1.46 (KHTML, like Gecko)Version/9.0 Mobile/13B143 Safari/601.1 (compatible; Baiduspider-render/2.0;Smartapp; +http://www.baidu.com/search/spider.html)")
				req.Header.Set("X-Forwarded-For", "111.206.198.69")
				return req
			},
			wantCode: 200,
		},
		{
			name: "必应 - bingbot",
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/test", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Set("User-Agent", "Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm) Chrome/")
				req.Header.Set("X-Forwarded-For", "157.55.39.1")
				return req
			},
			wantCode: 200,
		},
		{
			name: "必应 - adidxbot",
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/test", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Set("User-Agent", "Mozilla/5.0 (Windows Phone 8.1; ARM; Trident/7.0; Touch; rv:11.0; IEMobile/11.0; NOKIA; Lumia 530) like Gecko (compatible; adidxbot/2.0; +http://www.bing.com/bingbot.htm)")
				req.Header.Set("X-Forwarded-For", "157.55.39.1")
				return req
			},
			wantCode: 200,
		},
		{
			name: "必应 - MicrosoftPreview",
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/test", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Set("User-Agent", "Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; MicrosoftPreview/2.0; +https://aka.ms/MicrosoftPreview) Chrome/W.X.Y.Z Safari/537.36")
				req.Header.Set("X-Forwarded-For", "157.55.39.1")
				return req
			},
			wantCode: 200,
		},
		{
			name: "谷歌 - Googlebot",
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/test", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 6.0.1; Nexus 5X Build/MMB29P) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/W.X.Y.Z Mobile Safari/537.36 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")
				req.Header.Set("X-Forwarded-For", "66.249.66.1")
				return req
			},
			wantCode: 200,
		},
		{
			name: "谷歌 - Googlebot-Image",
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/test", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 6.0.1; Nexus 5X Build/MMB29P) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/W.X.Y.Z Mobile Safari/537.36 (compatible; Googlebot-Image/1.0; +http://www.google.com/bot.html)")
				req.Header.Set("X-Forwarded-For", "35.247.243.240")
				return req
			},
			wantCode: 200,
		},
		{
			name: "谷歌 - Googlebot-News",
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/test", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 6.0.1; Nexus 5X Build/MMB29P) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/W.X.Y.Z Mobile Safari/537.36 (compatible; Googlebot-News/1.0; +http://www.google.com/bot.html)")
				req.Header.Set("X-Forwarded-For", "66.249.90.77")
				return req
			},
			wantCode: 200,
		},
		{
			name: "谷歌 - Storebot-Google",
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/test", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; Storebot-Google/1.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/W.X.Y.Z Safari/537.36")
				req.Header.Set("X-Forwarded-For", "66.249.66.1")
				return req
			},
			wantCode: 200,
		},
		{
			name: "谷歌 - Google-InspectionTool",
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/test", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 6.0.1; Nexus 5X Build/MMB29P) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/W.X.Y.Z Mobile Safari/537.36 (compatible; Google-InspectionTool/1.0;)")
				req.Header.Set("X-Forwarded-For", "66.249.66.1")
				return req
			},
			wantCode: 200,
		},
		{
			name: "谷歌 - GoogleOther",
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/test", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; GoogleOther/1.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/W.X.Y.Z Safari/537.36")
				req.Header.Set("X-Forwarded-For", "66.249.66.1")
				return req
			},
			wantCode: 200,
		},
		{
			name: "谷歌 - Google-Extended",
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/test", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; Google-Extended/1.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/W.X.Y.Z Safari/537.36")
				req.Header.Set("X-Forwarded-For", "66.249.66.1")
				return req
			},
			wantCode: 200,
		},
		{
			name: "搜狗 - Sogou web spider",
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/test", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; Google-Extended/1.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/W.X.Y.Z Safari/537.36")
				req.Header.Set("X-Forwarded-For", "66.249.66.1")
				return req
			},
			wantCode: 200,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := gin.Default()
			server.TrustedPlatform = "X-Forwarded-For"
			server.Use(NewBuilder().Build())
			server.GET("/test", func(ctx *gin.Context) {
				ctx.JSON(200, nil)
			})

			recorder := httptest.NewRecorder()
			req := tc.reqBuilder(t)

			server.ServeHTTP(recorder, req)

			require.Equal(t, tc.wantCode, recorder.Code)
		})
	}
}

func TestBuilder_AddUserAgent(t *testing.T) {
	b := NewBuilder().AddUserAgent(map[string][]string{
		Baidu: {"test-new-baidu-user-agent"},
	})
	v, exist := b.crawlersMap["test-new-baidu-user-agent"]
	require.Equal(t, Baidu, v)
	require.True(t, exist)
}

func TestBuilder_RemoveUserAgent(t *testing.T) {
	b := NewBuilder().RemoveUserAgent("Baiduspider")
	v, exist := b.crawlersMap["Baiduspider"]
	require.Equal(t, "", v)
	require.False(t, exist)
}
