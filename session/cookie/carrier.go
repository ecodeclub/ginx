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

package cookie

import (
	"github.com/ecodeclub/ginx/gctx"
	"github.com/ecodeclub/ginx/session"
)

var _ session.TokenCarrier = &TokenCarrier{}

type TokenCarrier struct {
	MaxAge   int
	Name     string
	Path     string
	Domain   string
	Secure   bool
	HttpOnly bool
}

func (t *TokenCarrier) Clear(ctx *gctx.Context) {
	// 当 MaxAge 等于 -1 的时候，等价于清除 cookie
	ctx.SetCookie(t.Name, "", -1, t.Path, t.Domain, t.Secure, t.HttpOnly)
}

func (t *TokenCarrier) Inject(ctx *gctx.Context, value string) {
	ctx.SetCookie(t.Name, value, t.MaxAge, t.Path, t.Domain, t.Secure, t.HttpOnly)
}

func (t *TokenCarrier) Extract(ctx *gctx.Context) string {
	return ctx.Cookie(t.Name).StringOrDefault("")
}
