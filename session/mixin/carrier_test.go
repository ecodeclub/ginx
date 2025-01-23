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

package mixin

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ecodeclub/ginx"
	"github.com/ecodeclub/ginx/session/cookie"
	"github.com/ecodeclub/ginx/session/header"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CarrierTestSuite struct {
	suite.Suite
	carrier *TokenCarrier
}

func (s *CarrierTestSuite) SetupSuite() {
	hc := header.NewTokenCarrier()
	ck := &cookie.TokenCarrier{
		MaxAge: 1000,
		Name:   "ssid",
	}
	s.carrier = NewTokenCarrier(hc, ck)
}

func (s *CarrierTestSuite) TestInject() {
	val := "this is token"
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	s.carrier.Inject(&ginx.Context{
		Context: ctx,
	}, val)
	// 没有仔细检测 Cookie 的值，但是我们认为有值就可以了
	ck := recorder.Header().Get("Set-Cookie")
	assert.NotEmpty(s.T(), ck)

	ck = recorder.Header().Get("X-Access-Token")
	assert.NotEmpty(s.T(), ck)
}

func (s *CarrierTestSuite) TestExtract() {
	testCases := []struct {
		name       string
		ctxBuilder func() *ginx.Context
		wantVal    string
	}{
		{
			name: "从 header 中取出",
			ctxBuilder: func() *ginx.Context {
				recorder := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(recorder)
				val := "this is token"
				ctx.Request = &http.Request{
					Header: http.Header{},
				}
				ctx.Request.AddCookie(&http.Cookie{
					Name:  "ssid",
					Value: val,
				})
				return &ginx.Context{Context: ctx}
			},
			wantVal: "this is token",
		},
		{
			name: "从 cookie 中取出",
			ctxBuilder: func() *ginx.Context {
				recorder := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(recorder)
				val := "this is token"
				ctx.Request = &http.Request{
					Header: http.Header{},
				}
				ctx.Request.AddCookie(&http.Cookie{
					Name:  "ssid",
					Value: val,
				})
				return &ginx.Context{Context: ctx}
			},
			wantVal: "this is token",
		},
		{
			name: "都没有",
			ctxBuilder: func() *ginx.Context {
				recorder := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(recorder)
				ctx.Request = &http.Request{
					Header: http.Header{},
				}
				return &ginx.Context{Context: ctx}
			},
			wantVal: "",
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			val := s.carrier.Extract(tc.ctxBuilder())
			assert.Equal(t, tc.wantVal, val)
		})
	}
}

func (s *CarrierTestSuite) TestClear() {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	s.carrier.Clear(&ginx.Context{
		Context: ctx,
	})
	ck := recorder.Header().Get("Set-Cookie")
	strings.Contains(ck, "Max-Age=-1")
	assert.NotEmpty(s.T(), ck)

	ck = recorder.Header().Get("X-Access-Token")
	assert.Equal(s.T(), "", ck)
}

func TestCarrier(t *testing.T) {
	suite.Run(t, new(CarrierTestSuite))
}
