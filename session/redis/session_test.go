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

//go:build e2e

package redis

import (
	"context"
	"testing"
	"time"

	"github.com/ecodeclub/ginx/internal/e2e"
	"github.com/ecodeclub/ginx/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type SessionE2ETestSuite struct {
	e2e.BaseSuite
}

func (s *SessionE2ETestSuite) TestGetSetDel() {
	ssid := "test_ssid"
	sess := newRedisSession(ssid, time.Minute, s.RDB, session.Claims{
		Uid:  123,
		SSID: ssid,
		Data: map[string]string{
			"key1": "value1",
		},
	})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	defer sess.Destroy(ctx)
	ssKey1, ssVal1 := "ss_key1", "ss_val1"
	err := sess.Set(ctx, ssKey1, ssVal1)
	require.NoError(s.T(), err)
	val, err := sess.Get(ctx, ssKey1).AsString()
	require.NoError(s.T(), err)
	assert.Equal(s.T(), ssVal1, val)
}

func TestSession(t *testing.T) {
	suite.Run(t, new(SessionE2ETestSuite))
}
