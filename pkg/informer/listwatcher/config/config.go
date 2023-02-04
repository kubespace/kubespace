package config

import (
	"github.com/go-redis/redis/v8"
	"github.com/kubespace/kubespace/pkg/core/datatype"
	"github.com/kubespace/kubespace/pkg/core/db"
	"github.com/kubespace/kubespace/pkg/informer/listwatcher/storage"
	"gorm.io/gorm"
)

type ListWatcherConfig struct {
	DB          *gorm.DB
	RedisClient *redis.Client
	ResyncSec   int
}

func NewListWatcherConfig(db *db.DB, resyncSec int) *ListWatcherConfig {
	if resyncSec <= 0 {
		resyncSec = 5
	}
	return &ListWatcherConfig{
		DB:          db.Instance,
		RedisClient: db.RedisInstance,
		ResyncSec:   resyncSec,
	}
}

func (c *ListWatcherConfig) NewStorage(
	watchKey string,
	listFunc storage.ListFunc,
	filterFunc storage.FilterFunc,
	pResyncSec *int,
	dataType datatype.DataType) storage.Storage {
	resyncSec := c.ResyncSec
	if pResyncSec != nil {
		resyncSec = *pResyncSec
	}
	return storage.NewRedisStorage(c.RedisClient, watchKey, listFunc, filterFunc, resyncSec, dataType)
}
