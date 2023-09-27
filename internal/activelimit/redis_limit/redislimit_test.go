package redis_limit

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestNewRedisLimit(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:16379",
		Password: "",
		DB:       0,
	})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	err := redisClient.Ping(ctx).Err()
	if err != nil {
		panic("redis  连接失败")
	}
	testCases := []struct {
		name        string
		createLimit func() *RedisLimit
		before      func(localLimit *RedisLimit, key string, maxCount int64)

		maxCount   int64
		key        string
		wantBool   bool
		wantErr    error
		interval   time.Duration
		after      func(localLimit *RedisLimit, string2 string) (int64, error)
		afterCount int64
		afterErr   error
	}{
		{
			name: "正常通行没有退出",
			createLimit: func() *RedisLimit {
				return NewRedisLimit(redisClient)
			},
			before: func(localLimit *RedisLimit, key string, maxCount int64) {

			},
			maxCount: 1,
			key:      "test",
			wantBool: false,
			wantErr:  nil,
			after: func(localLimit *RedisLimit, key string) (int64, error) {
				defer func() {
					redisClient.Del(context.Background(), key)
				}()
				return redisClient.Get(context.Background(), key).Int64()
			},
			afterCount: 1,
			afterErr:   nil,
		},
		{
			name: "正常通行并退出",
			createLimit: func() *RedisLimit {
				return NewRedisLimit(redisClient)
			},
			before: func(localLimit *RedisLimit, key string, maxCount int64) {

			},
			maxCount: 1,
			key:      "test",
			wantBool: false,
			wantErr:  nil,
			after: func(localLimit *RedisLimit, key string) (int64, error) {
				defer func() {
					redisClient.Del(context.Background(), key)
				}()
				//退出
				limited, err := localLimit.Sub(context.Background(), key)
				assert.Equal(t, limited, true)
				assert.Equal(t, err, nil)
				return redisClient.Get(context.Background(), key).Int64()
			},
			afterCount: 0,
			afterErr:   nil,
		},
		{
			name: "被阻塞",
			createLimit: func() *RedisLimit {
				return NewRedisLimit(redisClient)
			},
			before: func(localLimit *RedisLimit, key string, maxCount int64) {
				limited, err := localLimit.Add(context.Background(), key, maxCount)
				assert.Equal(t, limited, false)
				assert.Equal(t, err, nil)
			},
			maxCount: 1,
			key:      "test",
			wantBool: true,
			wantErr:  nil,
			after: func(localLimit *RedisLimit, key string) (int64, error) {
				defer func() {
					redisClient.Del(context.Background(), key)
				}()
				//退出
				return redisClient.Get(context.Background(), key).Int64()
			},
			afterCount: 1,
			afterErr:   nil,
		},
		{
			name: "被阻塞,等别人出来了就可以进去了",
			createLimit: func() *RedisLimit {
				return NewRedisLimit(redisClient)
			},
			before: func(localLimit *RedisLimit, key string, maxCount int64) {
				limited, err := localLimit.Add(context.Background(), key, maxCount)
				assert.Equal(t, limited, false)
				assert.Equal(t, err, nil)
				go func() {
					time.Sleep(time.Millisecond * 10)
					limited, err = localLimit.Sub(context.Background(), key)
					assert.Equal(t, limited, true)
					assert.Equal(t, err, nil)
				}()
			},
			maxCount: 1,
			key:      "test",
			wantBool: false,
			wantErr:  nil,
			after: func(localLimit *RedisLimit, key string) (int64, error) {
				defer func() {
					redisClient.Del(context.Background(), key)
				}()
				//退出
				limited, err := localLimit.Sub(context.Background(), key)
				assert.Equal(t, limited, true)
				assert.Equal(t, err, nil)
				return redisClient.Get(context.Background(), key).Int64()
			},
			interval:   time.Millisecond * 20,
			afterCount: 0,
			afterErr:   nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			l := tc.createLimit()
			tc.before(l, tc.key, tc.maxCount)
			time.Sleep(tc.interval)
			limited, err := l.Add(context.Background(), tc.key, tc.maxCount)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantBool, limited)

			afterCount, err := tc.after(l, tc.key)
			assert.Equal(t, tc.afterCount, afterCount)
			assert.Equal(t, tc.afterErr, err)

		})
	}
}
