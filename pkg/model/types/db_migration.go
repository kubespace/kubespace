package types

import "time"

// DBMigration 数据库迁移表，在每次迭代时，有可能需要更新数据库表结构，那么需要做一些迁移的动作，比如增加字段或者数据迁移
// 该表会记录最新一次的迁移版本，下次如果有新的迁移版本，那么从该版本继续迁移
type DBMigration struct {
	Id         uint      `gorm:"primaryKey" json:"id"`
	Version    string    `gorm:"size:255" json:"version"` // 当前迁移的最新版本
	Lock       bool      `json:"lock"`                    // 迁移时首先要获取锁，保证每次只有一个实例在迁移
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
}

func (d DBMigration) TableName() string {
	return "db_migration"
}
