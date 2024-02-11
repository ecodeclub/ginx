package session

import (
	"context"

	"github.com/ecodeclub/ekit"
	"github.com/ecodeclub/ginx/gctx"
	"github.com/ecodeclub/ginx/internal/errs"
)

// Session 混合了 JWT 的设计。
type Session interface {
	// Set 将数据写入到 Session 里面
	Set(ctx context.Context, key string, val any) error
	// Get 从 Session 中获取数据，注意，这个方法不会从 JWT 里面获取数据
	Get(ctx context.Context, key string) ekit.AnyValue
	// Claims 编码进去了 JWT 里面的数据
	Claims() Claims
}

// Provider 定义了 Session 的整个管理机制。
// 所有的 Session 都必须支持 jwt
type Provider interface {
	// NewSession 将会初始化 Session
	// 其中 jwtData 将编码进去 jwt 中
	// sessData 将被放进去 Session 中
	NewSession(ctx *gctx.Context, uid int64, jwtData map[string]string,
		sessData map[string]any) (Session, error)
	// Get 尝试拿到 Session，如果没有，返回 error
	// Get 本身并不校验 Session 的有效性
	Get(ctx *gctx.Context) (Session, error)
}

type Claims struct {
	Uid  int64
	SSID string
	Data map[string]string
}

func (c Claims) Get(key string) ekit.AnyValue {
	val, ok := c.Data[key]
	if !ok {
		return ekit.AnyValue{Err: errs.ErrSessionKeyNotFound}
	}
	return ekit.AnyValue{Val: val}
}
