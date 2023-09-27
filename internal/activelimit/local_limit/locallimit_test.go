package local_limit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewLocalLimit(t *testing.T) {
	testCases := []struct {
		name        string
		createLimit func() *LocalLimit
		before      func(localLimit *LocalLimit, key string, maxCount int64)
		maxCount    int64
		key         string
		wantBool    bool
		wantErr     error
		interval    time.Duration
	}{
		{
			name: "正常操作",

			createLimit: func() *LocalLimit {
				return NewLocalLimit()
			},
			before: func(localLimit *LocalLimit, key string, maxCount int64) {
				//limited, err := localLimit.Add(context.Background(), key, maxCount)
				//assert.Equal(t, limited, false)
				//assert.Equal(t, err, nil)
			},
			maxCount: 1,
			key:      "test",
			wantBool: false,
			wantErr:  nil,
		},
		{
			name: "请求过多了",

			createLimit: func() *LocalLimit {
				return NewLocalLimit()
			},
			before: func(localLimit *LocalLimit, key string, maxCount int64) {
				limited, err := localLimit.Add(context.Background(), key, maxCount)
				assert.Equal(t, limited, false)
				assert.Equal(t, err, nil)
			},
			maxCount: 1,
			key:      "test",
			wantBool: true,
			wantErr:  nil,
		},
		{
			name: "请求过多了,等待别人退出",

			createLimit: func() *LocalLimit {
				return NewLocalLimit()
			},
			before: func(localLimit *LocalLimit, key string, maxCount int64) {
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
			interval: time.Millisecond * 20,
			maxCount: 1,
			key:      "test",
			wantBool: false,
			wantErr:  nil,
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
		})
	}
}
