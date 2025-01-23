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

package mixin

import (
	"github.com/ecodeclub/ginx/gctx"
	"github.com/ecodeclub/ginx/session"
)

type TokenCarrier struct {
	carriers []session.TokenCarrier
}

func NewTokenCarrier(carriers ...session.TokenCarrier) *TokenCarrier {
	return &TokenCarrier{carriers: carriers}
}

func (t *TokenCarrier) Inject(ctx *gctx.Context, value string) {
	for _, carrier := range t.carriers {
		carrier.Inject(ctx, value)
	}
}

func (t *TokenCarrier) Extract(ctx *gctx.Context) string {
	for _, carrier := range t.carriers {
		val := carrier.Extract(ctx)
		if val != "" {
			return val
		}
	}
	return ""
}

func (t *TokenCarrier) Clear(ctx *gctx.Context) {
	for _, carrier := range t.carriers {
		carrier.Clear(ctx)
	}
}
