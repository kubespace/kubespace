package mysql

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/model/types"
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
	err = DbMigrate(db)
	if err != nil {
		return nil, err
	}
	return db, err
}

func DbMigrate(db *gorm.DB) error {
	var err error
	migrateTypes := []interface{}{
		&types.User{},
		&types.PipelineWorkspace{},
		&types.Pipeline{},
		&types.PipelineStage{},
		&types.PipelinePlugin{},
		&types.PipelineRun{},
		&types.PipelineRunStage{},
		&types.PipelineRunJob{},
		&types.PipelineRunJobLog{},

		&types.SettingsSecret{},
		&types.SettingsImageRegistry{},

		&types.Project{},
		&types.ProjectApp{},
		&types.AppVersion{},
	}
	for _, model := range migrateTypes {
		err = db.AutoMigrate(model)
		if err != nil {
			return err
		}
	}

	return nil
}
