//go:build !k8s

package config

var Config = config{
	Redis: RedisConfig{
		Addr: "localhost:16379",
	},
}
