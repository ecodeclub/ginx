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
	"github.com/ecodeclub/ginx/gctx"
	"github.com/gin-gonic/gin"
)

const CtxSessionKey = "_session"

var defaultProvider Provider

func NewSession(ctx *gctx.Context, uid int64,
	jwtData map[string]string,
	sessData map[string]any) (Session, error) {
	return defaultProvider.NewSession(
		ctx,
		uid,
		jwtData,
		sessData)
}

// Get 参考 defaultProvider.Get 的说明
func Get(ctx *gctx.Context) (Session, error) {
	return defaultProvider.Get(ctx)
}

func SetDefaultProvider(sp Provider) {
	defaultProvider = sp
}

func DefaultProvider() Provider {
	return defaultProvider
}

func CheckLoginMiddleware() gin.HandlerFunc {
	return (&MiddlewareBuilder{sp: defaultProvider}).Build()
}

func RenewAccessToken(ctx *gctx.Context) error {
	return defaultProvider.RenewAccessToken(ctx)
}

func UpdateClaims(ctx *gctx.Context, claims Claims) error {
	return defaultProvider.UpdateClaims(ctx, claims)
}
