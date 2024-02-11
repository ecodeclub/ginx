package session

import (
	"context"
	"testing"

	"github.com/ecodeclub/ekit"
	"github.com/ecodeclub/ginx/internal/errs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemorySession_GetSet(t *testing.T) {
	testCases := []struct {
		name string

		// 插入数据
		key string
		val string

		getKey string

		wantVal ekit.AnyValue
	}{
		{
			name:    "成功获取",
			key:     "key1",
			val:     "value1",
			getKey:  "key1",
			wantVal: ekit.AnyValue{Val: "value1"},
		},
		{
			name:    "没有数据",
			key:     "key1",
			val:     "value1",
			getKey:  "key2",
			wantVal: ekit.AnyValue{Err: errs.ErrSessionKeyNotFound},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ms := NewMemorySession(Claims{})
			ctx := context.Background()
			err := ms.Set(ctx, tc.key, tc.val)
			require.NoError(t, err)
			val := ms.Get(ctx, tc.getKey)
			assert.Equal(t, tc.wantVal, val)
		})
	}
}
