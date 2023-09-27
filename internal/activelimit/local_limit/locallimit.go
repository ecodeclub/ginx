package local_limit

import (
	"context"

	"go.uber.org/atomic"
)

type LocalLimit struct {
	countActive *atomic.Int64
	//这里可以增加字段 以及对应的字段进行 分别限流
	//先判断总数  再判断单个限流目标的数量就行
}

func NewLocalLimit() *LocalLimit {
	return &LocalLimit{
		countActive: atomic.NewInt64(0),
	}
}

func (l *LocalLimit) Add(ctx context.Context, key string, maxCount int64) (bool, error) {
	// 并直接占坑成功
	if l.countActive.Add(1) <= maxCount {
		return false, nil
	} else {
		l.countActive.Sub(1)
		return true, nil
	}
}

func (l *LocalLimit) Sub(ctx context.Context, key string) (bool, error) {
	l.countActive.Sub(1)
	return true, nil
}
