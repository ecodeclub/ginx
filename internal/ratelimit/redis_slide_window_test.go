package ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestRedisSlidingWindowLimiter_Limit(t *testing.T) {
	r := NewRedisSlidingWindowLimiter(initRedis(), 500*time.Millisecond, 1)
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
