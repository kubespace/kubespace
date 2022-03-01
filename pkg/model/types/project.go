package types

import "time"

type Project struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:255;not null;uniqueIndex" json:"name"`
	Description string    `gorm:"size:2000;not null;" json:"description"`
	ClusterId   string    `gorm:"size:255;not null;" json:"cluster_id"`
	Namespace   string    `gorm:"size:255;not null;" json:"namespace"`
	Owner       string    `gorm:"size:255;not null" json:"owner"`
	CreateUser  string    `gorm:"size:255;not null" json:"create_user"`
	UpdateUser  string    `gorm:"size:255;not null" json:"update_user"`
	CreateTime  time.Time `gorm:"column:create_time;not null;autoCreateTime" json:"create_time"`
	UpdateTime  time.Time `gorm:"column:update_time;not null;autoUpdateTime" json:"update_time"`
}

const (
	AppStatusUninstall    = "UnInstall"
	AppStatusUnReady      = "UnReady"
	AppStatusRunningFault = "RunningFault"
	AppStatusRunning      = "Running"
)

type ProjectApp struct {
	ID           uint        `gorm:"primaryKey" json:"id"`
	ProjectId    uint        `gorm:"not null;uniqueIndex:ProjectNameUnique" json:"project_app_id"`
	Name         string      `gorm:"size:255;not null;uniqueIndex:ProjectNameUnique" json:"name"`
	AppVersionId uint        `gorm:"" json:"app_version_id"`
	AppVersion   *AppVersion `gorm:"-" json:"app_version"`
	Status       string      `gorm:"not null;size:255" json:"status"`
	CreateUser   string      `gorm:"size:255;not null" json:"create_user"`
	UpdateUser   string      `gorm:"size:255;not null" json:"update_user"`
	CreateTime   time.Time   `gorm:"column:create_time;not null;autoCreateTime" json:"create_time"`
	UpdateTime   time.Time   `gorm:"column:update_time;not null;autoUpdateTime" json:"update_time"`
}

type AppStore struct {
	ID   uint
	Name string
}

const (
	AppVersionScopeProjectApp = "project_app"
	AppVersionScopeStoreApp   = "store_app"

	AppVersionTypeOrdinaryApp = "ordinary_app"
	AppVersionTypeMiddleware  = "middleware"
	AppVersionTypeImportApp   = "import_app"
)

type AppVersion struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Scope          string    `gorm:"size:255;not null;uniqueIndex:ScopeAppNameVersionUnique" json:"scope"`
	ScopeId        uint      `gorm:"not null;uniqueIndex:ScopeAppNameVersionUnique" json:"scope_id"`
	PackageName    string    `gorm:"size:255;not null;uniqueIndex:ScopeAppNameVersionUnique" json:"package_name"`
	PackageVersion string    `gorm:"size:255;not null;uniqueIndex:ScopeAppNameVersionUnique" json:"package_version"`
	Type           string    `gorm:"size:255;not null;" json:"type"`
	AppVersion     string    `gorm:"size:255;not null" json:"app_version"`
	Values         string    `gorm:"type:text;not null" json:"values"`
	ChartPath      string    `gorm:"size:255;not null" json:"chart_path"`
	CreateUser     string    `gorm:"size:50;not null" json:"create_user"`
	CreateTime     time.Time `gorm:"column:create_time;not null;autoCreateTime" json:"create_time"`
	UpdateTime     time.Time `gorm:"column:update_time;not null;autoUpdateTime" json:"update_time"`
}

type AppVersionChart struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Path       string    `gorm:"size:255;uniqueIndex"`
	Content    []byte    `gorm:"type:mediumblob" json:"content"`
	CreateTime time.Time `gorm:"column:create_time;not null;autoCreateTime" json:"create_time"`
	UpdateTime time.Time `gorm:"column:update_time;not null;autoUpdateTime" json:"update_time"`
}
