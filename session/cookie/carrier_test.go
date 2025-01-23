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

package cookie

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ecodeclub/ginx"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CarrierTestSuite struct {
	suite.Suite
}

func (s *CarrierTestSuite) TestInject() {
	instance := &TokenCarrier{
		Name: "ssid",
	}
	val := "this is token"
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	instance.Inject(&ginx.Context{
		Context: ctx,
	}, val)
	// 没有仔细检测 Cookie 的值，但是我们认为有值就可以了
	ck := recorder.Header().Get("Set-Cookie")
	assert.NotEmpty(s.T(), ck)
}

func (s *CarrierTestSuite) TestExtract() {
	instance := &TokenCarrier{
		Name: "ssid",
	}
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
	res := instance.Extract(&ginx.Context{
		Context: ctx,
	})
	assert.Equal(s.T(), val, res)
}

func (s *CarrierTestSuite) TestClear() {
	instance := &TokenCarrier{
		Name: "ssid",
	}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	instance.Clear(&ginx.Context{
		Context: ctx,
	})
	ck := recorder.Header().Get("Set-Cookie")
	strings.Contains(ck, "Max-Age=-1")
	assert.NotEmpty(s.T(), ck)
}

func TestCarrier(t *testing.T) {
	suite.Run(t, new(CarrierTestSuite))
}
