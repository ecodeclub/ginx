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
