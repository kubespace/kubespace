package db

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type MysqlConfig struct {
	Username string
	Password string
	Host     string
	DbName   string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type Config struct {
	Mysql *MysqlConfig
	Redis *RedisConfig
}

type DB struct {
	config        *Config
	Instance      *gorm.DB
	RedisInstance *redis.Client
}

func NewDb(c *Config) (*DB, error) {
	mysqlInstance, err := NewMysqlDb(c.Mysql)
	if err != nil {
		return nil, err
	}
	return &DB{
		config:        c,
		Instance:      mysqlInstance,
		RedisInstance: NewRedisClient(c.Redis),
	}, nil
}

func NewMysqlDb(c *MysqlConfig) (db *gorm.DB, err error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", c.Username, c.Password, c.Host, c.DbName)
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return db, err
}

func NewRedisClient(c *RedisConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     c.Addr,
		Password: c.Password,
		DB:       c.DB,
	})
}
