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
	"net"
	"slices"
	"strings"
)

type SoGouStrategy struct {
	Hosts []string
}

func NewSoGouStrategy() *SoGouStrategy {
	return &SoGouStrategy{
		Hosts: []string{"sogou.com"},
	}
}

func (s *SoGouStrategy) CheckCrawler(ip string) (bool, error) {
	names, err := net.LookupAddr(ip)
	if err != nil {
		return false, err
	}
	if len(names) == 0 {
		return false, nil
	}
	return s.matchHost(names), nil
}

func (s *SoGouStrategy) matchHost(names []string) bool {
	return slices.ContainsFunc(s.Hosts, func(host string) bool {
		return slices.ContainsFunc(names, func(name string) bool {
			return strings.Contains(name, host)
		})
	})
}
