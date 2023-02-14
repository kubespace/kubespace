package migration

import (
	"fmt"
	"gorm.io/gorm"
)

const FirstVersion = ""

var Migrations []*Migration

// Register 注册迁移函数
func Register(m *Migration) {
	Migrations = append(Migrations, m)
}

// GetMigrationsMap 获取所有迁移版本的字典，key是当前版本指向的前一个版本
// 如果有两个指向一样的父版本号，则返回报错，保证不会分叉
func GetMigrationsMap() (map[string]*Migration, error) {
	var migrationsMap map[string]*Migration
	for _, m := range Migrations {
		if _, ok := migrationsMap[m.ParentVersion]; ok {
			return nil, fmt.Errorf("存在两个相同父版本号的迁移：version=%s，parent_version=%s", m.Version, m.ParentVersion)
		}
		migrationsMap[m.ParentVersion] = m
	}
	return migrationsMap, nil
}

// MigrateFunc 迁移函数
type MigrateFunc func(db *gorm.DB) error

type Migration struct {
	// 当前迁移的版本号
	Version string
	// 上一个版本号
	ParentVersion string
	// 真正的迁移方法
	MigrateFunc MigrateFunc
	// 迁移说明
	Description string
}
