package types

import "time"

const (
	AppStatusUninstall    = "UnInstall"
	AppStatusNotReady     = "NotReady"
	AppStatusRunningFault = "RunningFault"
	AppStatusRunning      = "Running"

	// AppTypeOrdinaryApp 普通应用
	AppTypeOrdinaryApp = "ordinary_app"
	// AppTypeMiddleware 中间件
	AppTypeMiddleware = "middleware"
	// AppTypeClusterComponent 集群组件
	AppTypeClusterComponent = "component"
)

type App struct {
	ID           uint        `gorm:"primaryKey" json:"id"`
	Scope        string      `gorm:"size:50;not null;uniqueIndex:ScopeNameUnique;comment:应用所属范围，包括project_app/store_app/component", json:"scope"`
	ScopeId      uint        `gorm:"not null;uniqueIndex:ScopeNameUnique" json:"scope_id"`
	Name         string      `gorm:"size:255;not null;uniqueIndex:ScopeNameUnique" json:"name"`
	Description  string      `gorm:"type:text;" json:"description"`
	AppVersionId uint        `gorm:"comment:当前最新或者正在运行的应用版本" json:"app_version_id"`
	AppVersion   *AppVersion `gorm:"-" json:"app_version"`
	Type         string      `gorm:"size:255;not null;comment:应用类型" json:"type"`
	Status       string      `gorm:"not null;size:255;comment:应用状态" json:"status"`
	Namespace    string      `gorm:"size:255;" json:"namespace"`
	CreateUser   string      `gorm:"size:255;not null" json:"create_user"`
	UpdateUser   string      `gorm:"size:255;not null" json:"update_user"`
	CreateTime   time.Time   `gorm:"column:create_time;not null;autoCreateTime" json:"create_time"`
	UpdateTime   time.Time   `gorm:"column:update_time;not null;autoUpdateTime" json:"update_time"`

	PodsNum      int `gorm:"-" json:"pods_num"`
	ReadyPodsNum int `gorm:"-" json:"ready_pods_num"`
}

// AppRevision 应用安装升级历史记录
type AppRevision struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	AppId         uint      `gorm:"not null;uniqueIndex:ProjectAppBuildRevisionUnique" json:"app_id"`
	BuildRevision uint      `gorm:"not null;uniqueIndex:ProjectAppBuildRevisionUnique" json:"build_revision"`
	AppVersionId  uint      `gorm:"not null;" json:"app_version_id"`
	Values        string    `gorm:"type:longtext;not null" json:"values"`
	CreateUser    string    `gorm:"size:50;not null" json:"create_user"`
	CreateTime    time.Time `gorm:"column:create_time;not null;autoCreateTime" json:"create_time"`
	UpdateTime    time.Time `gorm:"column:update_time;not null;autoUpdateTime" json:"update_time"`
}

// AppStore 应用商店应用
type AppStore struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:255;not null;" json:"name"`
	Description string    `gorm:"type:text;" json:"description"`
	Type        string    `gorm:"size:255;not null;comment:应用类型，包括普通应用/中间件/集群组件" json:"type"`
	Icon        []byte    `gorm:"type:mediumblob" json:"icon"`
	CreateUser  string    `gorm:"size:255;not null" json:"create_user"`
	UpdateUser  string    `gorm:"size:255;not null" json:"update_user"`
	CreateTime  time.Time `gorm:"column:create_time;not null;autoCreateTime" json:"create_time"`
	UpdateTime  time.Time `gorm:"column:update_time;not null;autoUpdateTime" json:"update_time"`
}

const (
	// AppVersionFromImport 导入应用
	AppVersionFromImport = "import"
	// AppVersionFromSpace 创建应用
	AppVersionFromSpace = "space"
)

// AppVersion 应用版本
type AppVersion struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Scope          string    `gorm:"size:255;not null;uniqueIndex:ScopeAppNameVersionUnique;comment:所属范围，工作空间应用/应用商店/集群组件" json:"scope"`
	ScopeId        uint      `gorm:"not null;uniqueIndex:ScopeAppNameVersionUnique;comment:工作空间应用id或应用商店id" json:"scope_id"`
	PackageName    string    `gorm:"size:255;not null;uniqueIndex:ScopeAppNameVersionUnique" json:"package_name"`
	PackageVersion string    `gorm:"size:255;not null;uniqueIndex:ScopeAppNameVersionUnique" json:"package_version"`
	From           string    `gorm:"size:255;not null;comment:应用版本来源" json:"from"`
	AppVersion     string    `gorm:"size:255;not null" json:"app_version"`
	Values         string    `gorm:"type:longtext;not null" json:"values"`
	Description    string    `gorm:"type:text;" json:"description"`
	ChartPath      string    `gorm:"size:255;not null;comment:该应用版本chart存储路径" json:"chart_path"`
	CreateUser     string    `gorm:"size:50;not null" json:"create_user"`
	CreateTime     time.Time `gorm:"column:create_time;not null;autoCreateTime" json:"create_time"`
	UpdateTime     time.Time `gorm:"column:update_time;not null;autoUpdateTime" json:"update_time"`
}

// AppVersionChart 应用版本chart存储，相同chart只存储一份
type AppVersionChart struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Path       string    `gorm:"size:255;uniqueIndex;comment:chart路径"`
	Content    []byte    `gorm:"type:mediumblob;comment:chart内容" json:"content"`
	CreateTime time.Time `gorm:"column:create_time;not null;autoCreateTime" json:"create_time"`
	UpdateTime time.Time `gorm:"column:update_time;not null;autoUpdateTime" json:"update_time"`
}
