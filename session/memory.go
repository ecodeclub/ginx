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
	"context"

	"github.com/ecodeclub/ginx/gctx"

	"github.com/ecodeclub/ekit"
	"github.com/ecodeclub/ginx/internal/errs"
)

var _ Session = &MemorySession{}

// MemorySession 一般用于测试
type MemorySession struct {
	data   map[string]any
	claims Claims
}

func (m *MemorySession) Destroy(ctx context.Context) error {
	return nil
}

func (m *MemorySession) UpdateClaims(ctx *gctx.Context, claims Claims) error {
	return nil
}

func (m *MemorySession) Del(ctx context.Context, key string) error {
	delete(m.data, key)
	return nil
}

func NewMemorySession(cl Claims) *MemorySession {
	return &MemorySession{
		data:   map[string]any{},
		claims: cl,
	}
}

func (m *MemorySession) Set(ctx context.Context, key string, val any) error {
	m.data[key] = val
	return nil
}

func (m *MemorySession) Get(ctx context.Context, key string) ekit.AnyValue {
	val, ok := m.data[key]
	if !ok {
		return ekit.AnyValue{Err: errs.ErrSessionKeyNotFound}
	}
	return ekit.AnyValue{Val: val}
}

func (m *MemorySession) Claims() Claims {
	return m.claims
}
