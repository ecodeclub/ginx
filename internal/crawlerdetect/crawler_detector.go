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

const (
	Baidu  = "baidu"
	Bing   = "bing"
	Google = "google"
	Sogou  = "sogou"
)

var strategyMap = map[string]Strategy{
	Baidu:  NewBaiduStrategy(),
	Bing:   NewBingStrategy(),
	Google: NewGoogleStrategy(),
	Sogou:  NewSoGouStrategy(),
}

type Strategy interface {
	CheckCrawler(ip string) (bool, error)
}

type UniversalStrategy struct {
	Hosts []string
}

func NewUniversalStrategy(hosts []string) *UniversalStrategy {
	return &UniversalStrategy{
		Hosts: hosts,
	}
}

func (s *UniversalStrategy) CheckCrawler(ip string) (bool, error) {
	names, err := net.LookupAddr(ip)
	if err != nil {
		return false, err
	}
	if len(names) == 0 {
		return false, nil
	}

	name, matched := s.matchHost(names)
	if !matched {
		return false, nil
	}

	ips, err := net.LookupIP(name)
	if err != nil {
		return false, err
	}
	if slices.ContainsFunc(ips, func(netIp net.IP) bool {
		return netIp.String() == ip
	}) {
		return true, nil
	}

	return false, nil
}

func (s *UniversalStrategy) matchHost(names []string) (string, bool) {
	var matchedName string
	return matchedName, slices.ContainsFunc(s.Hosts, func(host string) bool {
		return slices.ContainsFunc(names, func(name string) bool {
			if strings.Contains(name, host) {
				matchedName = name
				return true
			}
			return false
		})
	})
}

func NewCrawlerDetector(crawler string) Strategy {
	return strategyMap[crawler]
}
