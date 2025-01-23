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

package header

import (
	"strings"

	"github.com/ecodeclub/ginx/gctx"
	"github.com/ecodeclub/ginx/session"
)

type TokenCarrier struct {
	// 写入到 resp 中的名字
	// 固定从请求的 Authorization 字段中读取 token，并且假定使用的是 Bearer
	Name string
}

func (t *TokenCarrier) Clear(ctx *gctx.Context) {
	// 设置一个空的 token 就等价于清除了 token
	ctx.Writer.Header().Set(t.Name, "")
}

func (t *TokenCarrier) Inject(ctx *gctx.Context, value string) {
	ctx.Writer.Header().Set(t.Name, value)
}

// Extract 固定从 Authorization 中提取
func (t *TokenCarrier) Extract(ctx *gctx.Context) string {
	token := ctx.Request.Header.Get("Authorization")
	const bearerPrefix = "Bearer "
	return strings.TrimPrefix(token, bearerPrefix)
}

var _ session.TokenCarrier = &TokenCarrier{}

func NewTokenCarrier() *TokenCarrier {
	return &TokenCarrier{
		Name: "X-Access-Token",
	}
}
