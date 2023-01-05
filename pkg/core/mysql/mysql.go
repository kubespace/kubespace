package mysql

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Options struct {
	Username string
	Password string
	Host     string
	DbName   string
}

func NewMysqlDb(options *Options) (db *gorm.DB, err error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", options.Username, options.Password, options.Host, options.DbName)
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return db, err
}
