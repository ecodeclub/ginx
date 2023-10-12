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

package ratelimit

import (
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/ecodeclub/ginx/internal/ratelimit"
)

// NewRedisSlidingWindowLimiter 创建一个基于 redis 的滑动窗口限流器.
// cmd: 可传入 redis 的客户端
// interval: 窗口大小
// rate: 阈值
// 表示: 在 interval 内允许 rate 个请求
// 示例: 1s 内允许 3000 个请求
func NewRedisSlidingWindowLimiter(cmd redis.Cmdable,
	interval time.Duration, rate int) ratelimit.Limiter {
	return &ratelimit.RedisSlidingWindowLimiter{
		Cmd:      cmd,
		Interval: interval,
		Rate:     rate,
	}
}
