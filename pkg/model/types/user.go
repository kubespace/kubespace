package types

import "time"

const (
	ADMIN = "admin"
)

//type UserStore struct {
//	Common
//	Name      string `json:"name"`
//	Email     string `json:"email"`
//	Password  string `json:"password"`
//	Status    string `json:"status"`
//	IsSuper   string `json:"is_super"`
//	Roles     string `json:"roles"`
//	LastLogin string `json:"last_login"`
//}

type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Name     string `gorm:"size:255;not null;uniqueIndex" json:"name"`
	Email    string `gorm:"size:500" json:"email"`
	Password string `gorm:"size:1000;not null" json:"password"`
	Status   string `gorm:"size:255" json:"status"`
	IsSuper  bool   `json:"is_super"`
	//Roles     []string `json:"roles"`
	LastLogin  time.Time `json:"last_login"`
	CreateTime time.Time `gorm:"column:create_time;not null;autoCreateTime" json:"create_time"`
	UpdateTime time.Time `gorm:"column:update_time;not null;autoUpdateTime" json:"update_time"`
}

const (
	RoleScopePlatform = "platform"
	RoleScopeCluster  = "cluster"
	RoleScopePipeline = "pipeline"
	RoleScopeProject  = "project"

	RoleTypeViewer = "viewer"
	RoleTypeEditor = "editor"
	RoleTypeAdmin  = "admin"
)

type UserRole struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserId     uint      `gorm:"not null;uniqueIndex:idx_user_scope_id" json:"user_id"`
	UserName   string    `gorm:"-" json:"username"`
	Scope      string    `gorm:"size:50;not null;uniqueIndex:idx_user_scope_id" json:"scope"`
	ScopeId    uint      `gorm:"not null;uniqueIndex:idx_user_scope_id" json:"scope_id"`
	Role       string    `gorm:"size:50;not null;" json:"role"`
	CreateTime time.Time `gorm:"column:create_time;not null;autoCreateTime" json:"create_time"`
	UpdateTime time.Time `gorm:"column:update_time;not null;autoUpdateTime" json:"update_time"`
}
