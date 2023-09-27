package activelimit

import "context"

type Limiter interface {
	Add(ctx context.Context, key string, maxCount int64) (bool, error)
	Sub(ctx context.Context, key string) (bool, error)
}
