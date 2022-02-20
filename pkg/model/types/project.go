package types

import "time"

type Project struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"size:255;not null;uniqueIndex" json:"name"`
	Description         string    `gorm:"size:2000;not null;" json:"description"`
	ClusterId         string    `gorm:"size:255;not null;" json:"cluster_id"`
	Namespace         string    `gorm:"size:255;not null;" json:"namespace"`
	Owner   string    `gorm:"size:255;not null" json:"owner"`
	CreateUser   string    `gorm:"size:255;not null" json:"create_user"`
	UpdateUser   string    `gorm:"size:255;not null" json:"update_user"`
	CreateTime   time.Time `gorm:"column:create_time;not null;autoCreateTime" json:"create_time"`
	UpdateTime   time.Time `gorm:"column:update_time;not null;autoUpdateTime" json:"update_time"`
}

type ProjectApp struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	ProjectId uint `gorm:"not null;uniqueIndex:ProjectNameUnique" json:"project_app_id"`
	Name         string    `gorm:"size:255;not null;uniqueIndex:ProjectNameUnique" json:"name"`
	CurrentVersionId uint `gorm:"not null;" json:"current_version_id"`
	Status string `gorm:"not null;size:255" json:"status"`
	CreateUser   string    `gorm:"size:50;not null" json:"create_user"`
	CreateTime   time.Time `gorm:"column:create_time;not null;autoCreateTime" json:"create_time"`
	UpdateTime   time.Time `gorm:"column:update_time;not null;autoUpdateTime" json:"update_time"`
}

type AppStore struct {
	ID uint
	Name string
}

type AppVersion struct {
	ID uint `gorm:"primaryKey" json:"id"`
	ProjectAppId uint `gorm:"uniqueIndex:AppNameVersionUnique" json:"project_app_id"`
	AppStoreId uint `gorm:"uniqueIndex:AppNameVersionUnique" json:"app_store_id"`
	PackageName string `gorm:"size:255;not null;uniqueIndex:AppNameVersionUnique" json:"package_name"`
	PackageVersion string `gorm:"size:255;not null;uniqueIndex:AppNameVersionUnique" json:"package_version"`
	AppVersion string `gorm:"size:255;not null" json:"app_version"`
	Owner   string    `gorm:"not null;size:50;not null" json:"owner"`
	CreateUser   string    `gorm:"size:50;not null" json:"create_user"`
	CreateTime   time.Time `gorm:"column:create_time;not null;autoCreateTime" json:"create_time"`
	UpdateTime   time.Time `gorm:"column:update_time;not null;autoUpdateTime" json:"update_time"`
}
