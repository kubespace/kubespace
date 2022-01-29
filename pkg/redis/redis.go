package redis

import "github.com/go-redis/redis/v8"

type Options struct {
	Addr     string
	Password string
	DB       int
}

func NewRedisClient(op *Options) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     op.Addr,
		Password: op.Password,
		DB:       op.DB,
	})
}
