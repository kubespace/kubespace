package migrate

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/model/migrate/migration"
	"github.com/kubespace/kubespace/pkg/model/types"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
	"time"
)

type Migrate struct {
	db *gorm.DB
}

func NewMigrate(db *gorm.DB) *Migrate {
	return &Migrate{db: db}
}

// Do 执行数据库迁移
func (m *Migrate) Do() error {
	migrations, err := migration.GetMigrationsMap()
	if err != nil {
		return err
	}
	dbMigrate, err := m.dbLock()
	if err != nil {
		return err
	}
	defer func() {
		if err = m.dbUnlock(); err != nil {
			klog.Errorf("unlock db migration error: %s", err.Error())
		}
	}()
	if m.db.Migrator().HasTable("users") {
		// 存在user表，表示已经部署了业务，从数据库中记录的当前进行升级迁移
		return m.migrate(dbMigrate.Version, migrations)
	}
	// 还没有user表，表示第一次部署，直接初始化
	return m.init(m.getLastMigrateVersion(migrations))
}

// migrate 从指定的版本进行升级迁移
// 指定的版本是已经升级过的，从下一个版本开始执行迁移动作
func (m *Migrate) migrate(version string, migrations map[string]*migration.Migration) error {
	for mt, ok := migrations[version]; ok; {
		klog.Infof("start db migrate version=%s, description=%s", mt.Version, mt.Description)
		if err := mt.MigrateFunc(m.db); err != nil {
			klog.Errorf("db migrate error: %s", err.Error())
			return err
		}
		if err := m.db.Model(types.DBMigration{Id: 1}).Updates(
			&types.DBMigration{Version: mt.Version, UpdateTime: time.Now()}).Error; err != nil {
			klog.Errorf("update db migrate version error: %s", err.Error())
		}
		mt, ok = migrations[mt.Version]
	}
	return nil
}

// getLastMigrateVersion 获取最后一个版本号
func (m *Migrate) getLastMigrateVersion(migrations map[string]*migration.Migration) string {
	curr, ok := migrations[migration.FirstVersion]
	if !ok {
		return migration.FirstVersion
	}
	_, ok = migrations[curr.Version]
	// 是否有当前版本的下一个版本
	for ok {
		curr = migrations[curr.Version]
		_, ok = migrations[curr.Version]
	}
	return curr.Version
}

// init 第一次部署直接初始化数据库
func (m *Migrate) init(lastVersion string) error {
	if err := m.db.AutoMigrate(initTypes...); err != nil {
		return fmt.Errorf("init auto migrate error: %s", err.Error())
	}
	return m.db.Model(types.DBMigration{Id: 1}).Updates(
		&types.DBMigration{Version: lastVersion, UpdateTime: time.Now()}).Error
}

// dbLock 对db迁移进行加锁，若加锁失败，返回具体报错
func (m *Migrate) dbLock() (*types.DBMigration, error) {
	if err := m.db.AutoMigrate(&types.DBMigration{}); err != nil {
		return nil, err
	}
	one := &types.DBMigration{
		Id:         1,
		Lock:       false,
		Version:    migration.FirstVersion,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}
	// 查询id=1的迁移数据，若未找到，会初始化一条id=1的数据
	if err := m.db.FirstOrCreate(one, "id=1").Error; err != nil {
		return nil, err
	}
	// 如果返回lock=true，表示已经被锁，有其他实例在迁移数据
	if one.Lock {
		return nil, fmt.Errorf("db migration has locked")
	}
	one.Lock = true
	one.UpdateTime = time.Time{}
	// 乐观锁更新lock=true
	updates := m.db.Model(one).Where("lock=?", one.Lock).Select("lock", "update_time").Updates(one)
	if updates.Error != nil {
		return nil, updates.Error
	}
	if updates.RowsAffected == 0 {
		return nil, fmt.Errorf("not get db migration lock")
	}
	return one, nil
}

// dbUnlock db迁移解锁
func (m *Migrate) dbUnlock() error {
	if err := m.db.Model(types.DBMigration{Id: 1}).Updates(
		&types.DBMigration{Lock: false, UpdateTime: time.Now()}).Error; err != nil {
		return err
	}
	return nil
}
