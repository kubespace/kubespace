package config

import (
	"github.com/go-redis/redis/v8"
	"github.com/kubespace/kubespace/pkg/core/datatype"
	"github.com/kubespace/kubespace/pkg/core/db"
	storage2 "github.com/kubespace/kubespace/pkg/informer/listwatcher/storage"
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
	listFunc storage2.ListFunc,
	pResyncSec *int,
	dataType datatype.DataType) storage2.Storage {
	resyncSec := c.ResyncSec
	if pResyncSec != nil {
		resyncSec = *pResyncSec
	}
	return storage2.NewRedisStorage(c.RedisClient, watchKey, listFunc, resyncSec, dataType)
}
