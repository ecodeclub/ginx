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

package locallimit

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/atomic"
)

type LocalActiveLimit struct {
	// 最大限制个数
	maxActive *atomic.Int64
	// 当前活跃个数
	countActive *atomic.Int64
}

// NewLocalActiveLimit 全局限流
func NewLocalActiveLimit(maxActive int64) *LocalActiveLimit {
	return &LocalActiveLimit{
		maxActive:   atomic.NewInt64(maxActive),
		countActive: atomic.NewInt64(0),
	}
}

func (a *LocalActiveLimit) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		current := a.countActive.Add(1)
		defer func() {
			a.countActive.Sub(1)
		}()
		if current <= a.maxActive.Load() {
			ctx.Next()
		} else {
			ctx.AbortWithStatus(http.StatusTooManyRequests)
		}
	}
}
