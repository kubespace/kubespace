package types

import "k8s.io/klog"

const (
	OpGet    = "get"
	OpCreate = "create"
	OpUpdate = "update"
	OpDelete = "delete"
)

type RoleStore struct {
	Common
	Name        string `json:"name"`
	Description string `json:"description"`
	Permissions string `json:"permissions"`
}

type Role struct {
	Common
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Permissions []Permission `json:"permissions"`
}

type Permission struct {
	Scope      string   `json:"scope"`
	Object     string   `json:"object"`
	Name       string   `json:"name"`
	Operations []string `json:"operations"`
}

var AllPermissions = []Permission{
	{
		Scope:      "settings",
		Object:     "cluster",
		Name:       "集群管理",
		Operations: []string{OpGet, OpCreate, OpUpdate, OpDelete},
	},
	{
		Scope:      "settings",
		Object:     "user",
		Name:       "用户管理",
		Operations: []string{OpGet, OpCreate, OpUpdate, OpDelete},
	},
	{
		Scope:      "settings",
		Object:     "role",
		Name:       "角色管理",
		Operations: []string{OpGet, OpCreate, OpUpdate, OpDelete},
	},
	{
		Scope:      "cluster",
		Object:     "node",
		Name:       "节点管理",
		Operations: []string{OpGet, OpCreate, OpUpdate, OpDelete},
	},
	{
		Scope:      "cluster",
		Object:     "pod",
		Name:       "容器组",
		Operations: []string{OpGet, OpCreate, OpUpdate, OpDelete},
	},
	{
		Scope:      "cluster",
		Object:     "deployment",
		Name:       "无状态",
		Operations: []string{OpGet, OpCreate, OpUpdate, OpDelete},
	},
	{
		Scope:      "cluster",
		Object:     "statefulset",
		Name:       "有状态",
		Operations: []string{OpGet, OpCreate, OpUpdate, OpDelete},
	},
	{
		Scope:      "cluster",
		Object:     "daemonset",
		Name:       "守护进程集",
		Operations: []string{OpGet, OpCreate, OpUpdate, OpDelete},
	},
	{
		Scope:      "cluster",
		Object:     "job",
		Name:       "任务",
		Operations: []string{OpGet, OpCreate, OpUpdate, OpDelete},
	},
	{
		Scope:      "cluster",
		Object:     "cronjob",
		Name:       "定时任务",
		Operations: []string{OpGet, OpCreate, OpUpdate, OpDelete},
	},
	{
		Scope:      "cluster",
		Object:     "configmap",
		Name:       "配置项",
		Operations: []string{OpGet, OpCreate, OpUpdate, OpDelete},
	},
	{
		Scope:      "cluster",
		Object:     "secret",
		Name:       "保密字典",
		Operations: []string{OpGet, OpCreate, OpUpdate, OpDelete},
	},
	{
		Scope:      "cluster",
		Object:     "hpa",
		Name:       "水平扩缩容",
		Operations: []string{OpGet, OpCreate, OpUpdate, OpDelete},
	},
	{
		Scope:      "cluster",
		Object:     "service",
		Name:       "服务",
		Operations: []string{OpGet, OpCreate, OpUpdate, OpDelete},
	},
	{
		Scope:      "cluster",
		Object:     "ingress",
		Name:       "路由",
		Operations: []string{OpGet, OpCreate, OpUpdate, OpDelete},
	},
	{
		Scope:      "cluster",
		Object:     "networkPolicy",
		Name:       "网络策略",
		Operations: []string{OpGet, OpCreate, OpUpdate, OpDelete},
	},
	{
		Scope:      "cluster",
		Object:     "pvc",
		Name:       "存储声明",
		Operations: []string{OpGet, OpCreate, OpUpdate, OpDelete},
	},
	{
		Scope:      "cluster",
		Object:     "pv",
		Name:       "存储卷",
		Operations: []string{OpGet, OpCreate, OpUpdate, OpDelete},
	},
	{
		Scope:      "cluster",
		Object:     "sc",
		Name:       "存储类",
		Operations: []string{OpGet, OpCreate, OpUpdate, OpDelete},
	},
	{
		Scope:      "cluster",
		Object:     "namespace",
		Name:       "命名空间",
		Operations: []string{OpGet, OpCreate, OpUpdate, OpDelete},
	},
	{
		Scope:      "cluster",
		Object:     "event",
		Name:       "事件",
		Operations: []string{OpGet, OpCreate, OpUpdate, OpDelete},
	},
	{
		Scope:      "cluster",
		Object:     "serviceaccount",
		Name:       "服务账户",
		Operations: []string{OpGet, OpCreate, OpUpdate, OpDelete},
	},
	{
		Scope:      "cluster",
		Object:     "rolebinding",
		Name:       "角色绑定",
		Operations: []string{OpGet, OpCreate, OpUpdate, OpDelete},
	},
	{
		Scope:      "cluster",
		Object:     "role",
		Name:       "角色",
		Operations: []string{OpGet, OpCreate, OpUpdate, OpDelete},
	},
}

var (
	AdminRole = &Role{
		Name:        "admin",
		Description: "管理员角色，拥有所有对象权限",
		Permissions: AllPermissions,
	}

	EditRole = &Role{
		Name:        "edit",
		Description: "编辑角色，拥有集群对象的操作权限",
		Permissions: []Permission{},
	}

	ViewRole = &Role{
		Name:        "view",
		Description: "查看角色，拥有集群对象的查看权限，没有操作权限",
		Permissions: []Permission{},
	}
)

func init() {
	for _, p := range AllPermissions {
		if p.Scope == "cluster" {
			EditRole.Permissions = append(EditRole.Permissions, p)
			ViewRole.Permissions = append(ViewRole.Permissions, Permission{
				Name:       p.Name,
				Scope:      p.Scope,
				Object:     p.Object,
				Operations: []string{OpGet},
			})
		}
	}
	klog.Info("permission init")
}
