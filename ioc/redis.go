package ioc

import (
	"github.com/redis/go-redis/v9"

	"github.com/ecodeclub/ginx/config"
)

func InitRedis() redis.Cmdable {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.Config.Redis.Addr,
		Password: config.Config.Redis.Password,
	})
	return redisClient
}
