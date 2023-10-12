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

package accesslog

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuilder_Builder(t *testing.T) {
	testCases := []struct {
		name              string
		getReq            func() *http.Request
		accesslog         *AccessLog
		logfunc           func(accesslog *AccessLog) func(ctx context.Context, al *AccessLog)
		middleWarebuilder func(func(ctx context.Context, al *AccessLog)) gin.HandlerFunc
		setStatus         int
		setRsp            string
		resultAccessLog   *AccessLog
	}{
		{
			name: "不打印请求体,响应体",
			getReq: func() *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/accesslog", nil)
				require.NoError(t, err)
				return req
			},
			accesslog: &AccessLog{},
			logfunc: func(accesslog *AccessLog) func(ctx context.Context, al *AccessLog) {

				return func(ctx context.Context, al *AccessLog) {
					//
					//accesslog.Status = al.Status
					//accesslog.Method = al.Method
					////url 整个请求的u
					//accesslog.Url = al.Url
					////请求体
					//accesslog.ReqBody = al.ReqBody
					////响应体
					//accesslog.RespBody = al.RespBody
					////处理时间
					//accesslog.Duration = al.Duration
					////状态码
					//accesslog.Status = al.Status
					copy(accesslog, al)
					fmt.Printf("请求类型: %s \n请求url:%s \n请求体:%s \n响应体:%s \n状态码:%d \n消耗时间:%s \n", al.Method, al.Url, al.ReqBody, al.RespBody, al.Status, al.Duration)
				}
			},
			middleWarebuilder: func(f func(ctx context.Context, al *AccessLog)) gin.HandlerFunc {
				return NewBuilder(f).Builder()
			},
			resultAccessLog: &AccessLog{
				Method: "GET",
				Url:    "/accesslog",
			},
		},
		{
			name: "不打印请求体,打印响应体",
			getReq: func() *http.Request {
				req, err := http.NewRequest(http.MethodGet, "/accesslog", nil)
				require.NoError(t, err)
				return req
			},
			accesslog: &AccessLog{},
			logfunc: func(accesslog *AccessLog) func(ctx context.Context, al *AccessLog) {

				return func(ctx context.Context, al *AccessLog) {

					copy(accesslog, al)

					fmt.Printf("请求类型: %s \n请求url:%s \n请求体:%s \n响应体:%s \n状态码:%d \n消耗时间:%s \n", al.Method, al.Url, al.ReqBody, al.RespBody, al.Status, al.Duration)
				}
			},
			middleWarebuilder: func(f func(ctx context.Context, al *AccessLog)) gin.HandlerFunc {
				return NewBuilder(f).AllowRespBody().Builder()
			},
			resultAccessLog: &AccessLog{
				Method:   "GET",
				Url:      "/accesslog",
				RespBody: `{"msg":"aa22"}`,
				Status:   http.StatusOK,
			},
		},
		{
			name: "打印请求体,不打印响应体",
			getReq: func() *http.Request {
				read := strings.NewReader(`{"msg":"aa11"}`)

				req, err := http.NewRequest(http.MethodGet, "/accesslog", read)
				require.NoError(t, err)
				return req
			},
			accesslog: &AccessLog{},
			logfunc: func(accesslog *AccessLog) func(ctx context.Context, al *AccessLog) {

				return func(ctx context.Context, al *AccessLog) {

					copy(accesslog, al)

					fmt.Printf("请求类型: %s \n请求url:%s \n请求体:%s \n响应体:%s \n状态码:%d \n消耗时间:%s \n", al.Method, al.Url, al.ReqBody, al.RespBody, al.Status, al.Duration)
				}
			},
			middleWarebuilder: func(f func(ctx context.Context, al *AccessLog)) gin.HandlerFunc {
				return NewBuilder(f).AllowReqBody().Builder()
			},
			resultAccessLog: &AccessLog{
				Method:  "GET",
				Url:     "/accesslog",
				ReqBody: `{"msg":"aa11"}`,
			},
		},
		{
			name: "打印请求体,打印响应体",
			getReq: func() *http.Request {
				read := strings.NewReader(`{"msg":"aa11"}`)

				req, err := http.NewRequest(http.MethodGet, "/accesslog", read)
				require.NoError(t, err)
				return req
			},
			accesslog: &AccessLog{},
			logfunc: func(accesslog *AccessLog) func(ctx context.Context, al *AccessLog) {

				return func(ctx context.Context, al *AccessLog) {

					copy(accesslog, al)

					fmt.Printf("请求类型: %s \n请求url:%s \n请求体:%s \n响应体:%s \n状态码:%d \n消耗时间:%s \n", al.Method, al.Url, al.ReqBody, al.RespBody, al.Status, al.Duration)
				}
			},
			middleWarebuilder: func(f func(ctx context.Context, al *AccessLog)) gin.HandlerFunc {
				return NewBuilder(f).AllowReqBody().AllowRespBody().Builder()
			},
			resultAccessLog: &AccessLog{
				Method:   "GET",
				Url:      "/accesslog",
				ReqBody:  `{"msg":"aa11"}`,
				RespBody: `{"msg":"aa22"}`,
				Status:   http.StatusOK,
			},
		},
		{
			name: "打印请求体超标,不打印响应体,限制长度为10",
			getReq: func() *http.Request {
				read := strings.NewReader(`{"msg":"aa11"}`)

				req, err := http.NewRequest(http.MethodGet, "/accesslog", read)
				require.NoError(t, err)
				return req
			},
			accesslog: &AccessLog{},
			logfunc: func(accesslog *AccessLog) func(ctx context.Context, al *AccessLog) {

				return func(ctx context.Context, al *AccessLog) {

					copy(accesslog, al)

					fmt.Printf("请求类型: %s \n请求url:%s \n请求体:%s \n响应体:%s \n状态码:%d \n消耗时间:%s \n", al.Method, al.Url, al.ReqBody, al.RespBody, al.Status, al.Duration)
				}
			},
			middleWarebuilder: func(f func(ctx context.Context, al *AccessLog)) gin.HandlerFunc {
				return NewBuilder(f).AllowReqBody().MaxLength(10).Builder()
			},
			resultAccessLog: &AccessLog{
				Method:  "GET",
				Url:     "/accesslog",
				ReqBody: `{"msg":"aa`,
			},
		},
		{
			name: "不打印请求体,打印响应体超标,限制长度为10",
			getReq: func() *http.Request {
				read := strings.NewReader(`{"msg":"aa11"}`)

				req, err := http.NewRequest(http.MethodGet, "/accesslog", read)
				require.NoError(t, err)
				return req
			},
			accesslog: &AccessLog{},
			logfunc: func(accesslog *AccessLog) func(ctx context.Context, al *AccessLog) {

				return func(ctx context.Context, al *AccessLog) {

					copy(accesslog, al)

					fmt.Printf("请求类型: %s \n请求url:%s \n请求体:%s \n响应体:%s \n状态码:%d \n消耗时间:%s \n", al.Method, al.Url, al.ReqBody, al.RespBody, al.Status, al.Duration)
				}
			},
			middleWarebuilder: func(f func(ctx context.Context, al *AccessLog)) gin.HandlerFunc {
				return NewBuilder(f).AllowRespBody().MaxLength(10).Builder()
			},
			resultAccessLog: &AccessLog{
				Method:   "GET",
				Url:      "/accesslog",
				RespBody: `{"msg":"aa`,
				Status:   http.StatusOK,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := gin.Default()
			server.Use(tc.middleWarebuilder(tc.logfunc(tc.accesslog)))
			server.GET("/accesslog", func(ctx *gin.Context) {
				ctx.JSON(http.StatusOK, map[string]any{
					"msg": "aa22",
				})
			})
			resp := httptest.NewRecorder()

			server.ServeHTTP(resp, tc.getReq())
			//中间件使用的defer 所有这里要给点时间
			time.Sleep(time.Millisecond * 100)
			assert.Equal(t, tc.accesslog.Method, tc.resultAccessLog.Method)
			assert.Equal(t, tc.accesslog.Url, tc.resultAccessLog.Url)
			assert.Equal(t, tc.accesslog.ReqBody, tc.resultAccessLog.ReqBody)
			assert.Equal(t, tc.accesslog.RespBody, tc.resultAccessLog.RespBody)
			//时间不好判断
			//assert.Equal(t, tc.accesslog.Duration, tc.resultAccessLog.Duration)

			assert.Equal(t, tc.accesslog.Status, tc.resultAccessLog.Status)

		})

	}
}

func copy(source, target *AccessLog) {
	source.Status = target.Status
	source.Method = target.Method
	//url 整个请求的u
	source.Url = target.Url
	//请求体
	source.ReqBody = target.ReqBody
	//响应体
	source.RespBody = target.RespBody
	//处理时间
	source.Duration = target.Duration
	//状态码
	source.Status = target.Status
}
