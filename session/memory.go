package session

import (
	"context"

	"github.com/ecodeclub/ekit"
	"github.com/ecodeclub/ginx/internal/errs"
)

// MemorySession 一般用于测试
type MemorySession struct {
	data   map[string]any
	claims Claims
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
