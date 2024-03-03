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

package e2e

import (
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
)

type BaseSuite struct {
	suite.Suite
	RDB redis.Cmdable
}

func (s *BaseSuite) SetupSuite() {
	s.RDB = newRedisTestClient()
}

func (s *BaseSuite) TearDownSuite() {
	if s.RDB != nil {
		s.RDB.(*redis.Client).Close()
	}
}
