package config

import (
	"github.com/go-redis/redis/v8"
	"github.com/kubespace/kubespace/pkg/model/listwatcher/watch_storage"
	"gorm.io/gorm"
)

type ListWatcherConfig struct {
	DB          *gorm.DB
	RedisClient *redis.Client
}

func NewListWatcherConfig(db *gorm.DB, redisClient *redis.Client) *ListWatcherConfig {
	return &ListWatcherConfig{
		DB:          db,
		RedisClient: redisClient,
	}
}

func (c *ListWatcherConfig) NewWatchStorage(watchKey string) watch_storage.Storage {
	return watch_storage.NewRedisStorage(c.RedisClient, watchKey)
}
