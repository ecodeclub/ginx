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
	"log/slog"
	"net/http"
	"strings"

	"github.com/ecodeclub/ginx/internal/crawlerdetect"
	"github.com/gin-gonic/gin"
)

const Baidu = crawlerdetect.Baidu
const Bing = crawlerdetect.Bing
const Google = crawlerdetect.Google
const Sogou = crawlerdetect.Sogou

type Builder struct {
	crawlersMap map[string]string
}

func NewBuilder() *Builder {
	return &Builder{
		// 常用的 User-Agent 映射
		crawlersMap: map[string]string{
			"Baiduspider":        Baidu,
			"Baiduspider-render": Baidu,

			"bingbot":          Bing,
			"adidxbot":         Bing,
			"MicrosoftPreview": Bing,

			"Googlebot":             Google,
			"Googlebot-Image":       Google,
			"Googlebot-News":        Google,
			"Googlebot-Video":       Google,
			"Storebot-Google":       Google,
			"Google-InspectionTool": Google,
			"GoogleOther":           Google,
			"Google-Extended":       Google,

			"Sogou web spider": Sogou,
		},
	}
}

// AddUserAgent 添加 user-agent 映射
// 例如：
//
//	map[string][]string{
//		crawlerdetect.Baidu: []string{"NewBaiduUserAgent"},
//		crawlerdetect.Bing: []string{"NewBingUserAgent"},
//	}
func (b *Builder) AddUserAgent(userAgents map[string][]string) *Builder {
	for crawler, values := range userAgents {
		for _, userAgent := range values {
			b.crawlersMap[userAgent] = crawler
		}
	}
	return b
}

func (b *Builder) RemoveUserAgent(userAgents ...string) *Builder {
	for _, userAgent := range userAgents {
		delete(b.crawlersMap, userAgent)
	}
	return b
}

func (b *Builder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userAgent := ctx.GetHeader("User-Agent")
		ip := ctx.ClientIP()
		if ip == "" {
			slog.ErrorContext(ctx, "crawlerdetect", "error", "ip is empty.")
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}
		crawlerDetector := b.getCrawlerDetector(userAgent)
		if crawlerDetector == nil {
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}
		pass, err := crawlerDetector.CheckCrawler(ip)
		if err != nil {
			slog.ErrorContext(ctx, "crawlerdetect", "error", err.Error())
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if !pass {
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}
		ctx.Next()
	}
}

func (b *Builder) getCrawlerDetector(userAgent string) crawlerdetect.Strategy {
	for key, value := range b.crawlersMap {
		if strings.Contains(userAgent, key) {
			return crawlerdetect.NewCrawlerDetector(value)
		}
	}
	return nil
}
