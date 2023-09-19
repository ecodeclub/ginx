package config

type config struct {
	Redis RedisConfig
}

type RedisConfig struct {
	Addr     string
	Password string
}
