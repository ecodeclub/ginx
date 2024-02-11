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
	"testing"

	"github.com/ecodeclub/ginx/gctx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestBuilder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	p := NewMockProvider(ctrl)
	p.EXPECT().NewSession(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx *gctx.Context, uid int64, jwtData map[string]string,
			sessData map[string]any) (Session, error) {
			return &MemorySession{data: sessData,
				claims: Claims{Uid: uid, Data: jwtData}}, nil
		})
	sess, err := NewSessionBuilder(new(gctx.Context), 123).
		SetProvider(p).
		SetJwtData(map[string]string{"jwt": "true"}).
		SetSessData(map[string]any{"session": "true"}).
		Build()
	require.NoError(t, err)
	assert.Equal(t, &MemorySession{
		data: map[string]any{"session": "true"},
		claims: Claims{
			Uid:  123,
			Data: map[string]string{"jwt": "true"},
		},
	}, sess)
}
