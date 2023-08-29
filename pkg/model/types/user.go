package types

import "time"

const (
	ADMIN = "admin"
)

type User struct {
	ID         uint         `gorm:"primaryKey" json:"id"`
	Name       string       `gorm:"size:255;not null;uniqueIndex" json:"name"`
	Email      string       `gorm:"size:500" json:"email"`
	Password   string       `gorm:"size:1000;not null" json:"password"`
	Roles      *[]*UserRole `gorm:"-" json:"roles"`
	Status     string       `gorm:"size:255" json:"status"`
	IsSuper    bool         `json:"is_super"`
	LastLogin  time.Time    `json:"last_login"`
	CreateTime time.Time    `gorm:"column:create_time;not null;autoCreateTime" json:"create_time"`
	UpdateTime time.Time    `gorm:"column:update_time;not null;autoUpdateTime" json:"update_time"`
}

const (
	ScopePlatform = "platform"
	ScopeCluster  = "cluster"
	ScopePipeline = "pipeline"
	ScopeProject  = "project"
	ScopeAppStore = "appstore"

	// RoleViewer 观察员，范围内只有查询资源权限
	RoleViewer = "viewer"
	// RoleEditor 编辑员，范围内有查询、更新、创建、删除权限，但是没有成员管理以及删除该范围权限
	RoleEditor = "editor"
	// RoleAdmin 管理员，范围内所有权限
	RoleAdmin = "admin"
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
