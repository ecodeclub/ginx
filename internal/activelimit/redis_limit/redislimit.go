package redis_limit

import (
	"context"
	_ "embed"

	"github.com/redis/go-redis/v9"
)

//go:embed activeAdd.lua
var LuaActiveAdd string

//go:embed activeSub.lua
var LuaActiveSub string

type RedisLimit struct {
	cmd redis.Cmdable
}

func NewRedisLimit(cmd redis.Cmdable) *RedisLimit {
	return &RedisLimit{cmd: cmd}
}

func (r *RedisLimit) Add(ctx context.Context, key string, maxCount int64) (bool, error) {
	//TODO implement me
	return r.cmd.Eval(ctx, LuaActiveAdd, []string{key}, maxCount).Bool()
}

func (r *RedisLimit) Sub(ctx context.Context, key string) (bool, error) {
	//TODO implement me
	return r.cmd.Eval(ctx, LuaActiveSub, []string{key}).Bool()
}
