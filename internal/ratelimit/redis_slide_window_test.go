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

package ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestRedisSlidingWindowLimiter_Limit(t *testing.T) {
	r := &RedisSlidingWindowLimiter{
		Cmd:      initRedis(),
		Interval: 500 * time.Millisecond,
		Rate:     1,
	}
	tests := []struct {
		name     string
		ctx      context.Context
		key      string
		interval time.Duration
		want     bool
		wantErr  error
	}{
		{
			name: "正常通过",
			ctx:  context.Background(),
			key:  "foo",
			want: false,
		},
		{
			name: "另外一个key正常通过",
			ctx:  context.Background(),
			key:  "bar",
			want: false,
		},
		{
			name:     "限流",
			ctx:      context.Background(),
			key:      "foo",
			interval: 300 * time.Millisecond,
			want:     true,
		},
		{
			name:     "窗口有空余正常通过",
			ctx:      context.Background(),
			key:      "foo",
			interval: 510 * time.Millisecond,
			want:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			<-time.After(tt.interval)
			got, err := r.Limit(tt.ctx, tt.key)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func initRedis() redis.Cmdable {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:16379",
	})
	return redisClient
}

func TestRedisSlidingWindowLimiter(t *testing.T) {
	r := &RedisSlidingWindowLimiter{
		Cmd:      initRedis(),
		Interval: time.Second,
		Rate:     1200,
	}
	var (
		total      = 1500 // 总请求数
		succCount  int    // 成功请求数
		limitCount int    // 被限流的请求数
	)
	start := time.Now()
	for i := 0; i < total; i++ {
		limit, err := r.Limit(context.Background(), "TestRedisSlidingWindowLimiter")
		if err != nil {
			t.Fatalf("limit error: %v", err)
			return
		}
		if limit {
			limitCount++
			continue
		}
		succCount++
	}
	end := time.Now()
	t.Logf("开始时间: %v", start.Format(time.StampMilli))
	t.Logf("结束时间: %v", end.Format(time.StampMilli))
	t.Logf("total: %d, succ: %d, limited: %d", total, succCount, limitCount)
}
