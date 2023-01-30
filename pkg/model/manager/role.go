package manager

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/utils"
	"k8s.io/klog/v2"
	"strings"
)

type RoleManager struct {
	CommonManager
}

func NewRoleManager(redisClient *redis.Client) *RoleManager {
	return &RoleManager{
		CommonManager{
			client:   redisClient,
			modelKey: "osp:role",
			Context:  context.Background(),
		},
	}
}

func (r *RoleManager) Get(name string) (*types.Role, error) {
	roleObj := types.RoleStore{}
	if err := r.CommonManager.Get(name, &roleObj); err != nil {
		return nil, err
	}
	var perms []types.Permission
	json.Unmarshal([]byte(roleObj.Permissions), &perms)

	role := &types.Role{
		Name:        roleObj.Name,
		Description: roleObj.Description,
		Common:      roleObj.Common,
		Permissions: perms,
	}

	return role, nil
}

func (r *RoleManager) List(filters map[string]interface{}) ([]*types.Role, error) {
	dList, err := r.CommonManager.List(filters)
	if err != nil {
		return nil, err
	}
	jsonBody, err := json.Marshal(dList)
	if err != nil {
		return nil, err
	}
	var roleStores []*types.RoleStore

	if err := json.Unmarshal(jsonBody, &roleStores); err != nil {
		return nil, err
	}

	var roles []*types.Role
	for _, rs := range roleStores {
		var perms []types.Permission
		json.Unmarshal([]byte(rs.Permissions), &perms)

		roles = append(roles, &types.Role{
			Name:        rs.Name,
			Description: rs.Description,
			Common:      rs.Common,
			Permissions: perms,
		})
	}
	return roles, nil
}

func (r *RoleManager) Update(role *types.Role) error {
	permBytes, err := json.Marshal(role.Permissions)
	if err != nil {
		return err
	}
	roleStore := types.RoleStore{
		Name:        role.Name,
		Description: role.Description,
		Common:      role.Common,
		Permissions: string(permBytes),
	}
	if err := r.CommonManager.Save(role.Name, roleStore, 0, false); err != nil {
		return err
	}
	return nil
}

func (r *RoleManager) Create(role *types.Role) error {
	permBytes, err := json.Marshal(role.Permissions)
	if err != nil {
		return err
	}
	roleStore := types.RoleStore{
		Name:        role.Name,
		Description: role.Description,
		Common:      role.Common,
		Permissions: string(permBytes),
	}

	if err := r.CommonManager.Save(role.Name, roleStore, -1, true); err != nil {
		return err
	}
	return nil
}

func (r *RoleManager) InitRole(role *types.Role) error {
	_, err := r.Get(role.Name)
	if err != nil && strings.Contains(err.Error(), "not found key") {
		role.CreateTime = utils.StringNow()
		role.UpdateTime = utils.StringNow()
		return r.Create(role)
	} else if err != nil {
		return err
	}
	return nil
}

func (r *RoleManager) Init() {
	err := r.InitRole(types.AdminRole)
	if err != nil {
		klog.Error("init admin role error: ", err)
	}
	err = r.InitRole(types.EditRole)
	if err != nil {
		klog.Error("init edit role error: ", err)
	}
	err = r.InitRole(types.ViewRole)
	if err != nil {
		klog.Error("init view role error: ", err)
	}
}
