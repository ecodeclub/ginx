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

package crawlerdetect

import (
	"errors"
	"log"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBingStrategy(t *testing.T) {
	s := NewBingStrategy()
	require.NotNil(t, s)
	testCases := []struct {
		name    string
		ip      string
		matched bool
		errFunc require.ErrorAssertionFunc
	}{
		{
			name:    "无效 ip",
			ip:      "256.0.0.0",
			matched: false,
			errFunc: func(t require.TestingT, err error, i ...interface{}) {
				var dnsError *net.DNSError
				if !errors.As(err, &dnsError) {
					log.Fatal(err)
				}
			},
		},
		{
			name:    "非必应 ip",
			ip:      "166.249.90.77",
			matched: false,
		},
		{
			name:    "必应 ip",
			ip:      "157.55.39.1",
			matched: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := s.CheckCrawler(tc.ip)
			if err != nil {
				tc.errFunc(t, err)
			}
			require.Equal(t, tc.matched, m)
		})
	}
}
