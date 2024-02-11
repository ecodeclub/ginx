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

import "github.com/ecodeclub/ginx/gctx"

// Builder 是一个辅助接口，便于构造 Session
type Builder struct {
	ctx      *gctx.Context
	uid      int64
	jwtData  map[string]string
	sessData map[string]any
	sp       Provider
}

// NewSessionBuilder 创建一个 Builder 用于构造 Session
// 默认使用 defaultProvider
func NewSessionBuilder(ctx *gctx.Context, uid int64) *Builder {
	return &Builder{
		ctx: ctx,
		uid: uid,
		sp:  defaultProvider,
	}
}

func (b *Builder) SetProvider(p Provider) *Builder {
	b.sp = p
	return b
}

func (b *Builder) SetJwtData(data map[string]string) *Builder {
	b.jwtData = data
	return b
}

func (b *Builder) SetSessData(data map[string]any) *Builder {
	b.sessData = data
	return b
}

func (b *Builder) Build() (Session, error) {
	return b.sp.NewSession(b.ctx, b.uid, b.jwtData, b.sessData)
}
