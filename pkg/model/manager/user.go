package manager

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/utils"
	"strconv"
)

type UserManager struct {
	CommonManager
	role *RoleManager
}

func NewUserManager(redisClient *redis.Client, role *RoleManager) *UserManager {
	return &UserManager{
		CommonManager: CommonManager{
			client:   redisClient,
			modelKey: "osp:user",
			Context:  context.Background(),
		},
		role: role,
	}
}

func (u *UserManager) parseToStore(user *types.User) (*types.UserStore, error) {
	roleBytes, err := json.Marshal(user.Roles)
	if err != nil {
		return nil, err
	}
	userStore := &types.UserStore{
		Name:      user.Name,
		Email:     user.Email,
		Password:  user.Password,
		Status:    user.Status,
		IsSuper:   strconv.FormatBool(user.IsSuper),
		Roles:     string(roleBytes),
		LastLogin: user.LastLogin,
	}
	return userStore, nil
}

func (u *UserManager) parseToUser(userStore *types.UserStore) (*types.User, error) {
	var roles []string
	json.Unmarshal([]byte(userStore.Roles), &roles)

	user := &types.User{
		Name:      userStore.Name,
		Email:     userStore.Email,
		Common:    userStore.Common,
		Password:  userStore.Password,
		Status:    userStore.Status,
		IsSuper:   utils.ParseBool(userStore.IsSuper),
		LastLogin: userStore.LastLogin,
		Roles:     roles,
	}
	return user, nil
}

func (u *UserManager) Get(name string) (*types.User, error) {
	userObj := &types.UserStore{}
	if err := u.CommonManager.Get(name, userObj); err != nil {
		return nil, err
	}

	return u.parseToUser(userObj)
}

func (u *UserManager) List(filters map[string]interface{}) ([]*types.User, error) {
	dList, err := u.CommonManager.List(filters)
	if err != nil {
		return nil, err
	}
	jsonBody, err := json.Marshal(dList)
	if err != nil {
		return nil, err
	}
	var userStores []*types.UserStore

	if err := json.Unmarshal(jsonBody, &userStores); err != nil {
		return nil, err
	}

	var users []*types.User
	for _, rs := range userStores {
		user, err := u.parseToUser(rs)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (u *UserManager) Update(user *types.User) error {
	userStore, err := u.parseToStore(user)
	if err != nil {
		return err
	}
	if err := u.CommonManager.Save(user.Name, userStore, 0, false); err != nil {
		return err
	}
	return nil
}

func (u *UserManager) Create(user *types.User) error {
	userStore, err := u.parseToStore(user)
	if err != nil {
		return err
	}
	if err := u.CommonManager.Save(user.Name, userStore, -1, true); err != nil {
		return err
	}
	return nil
}

func (u *UserManager) Permissions(user *types.User) ([]types.Permission, error) {
	var perms []types.Permission

	for _, role := range user.Roles {
		roleObj, err := u.role.Get(role)
		if err != nil {
			return nil, err
		}
		for _, p := range roleObj.Permissions {
			has := false
			for _, perm := range perms {
				if p.Scope == perm.Scope && p.Object == perm.Object {
					has = true
					for _, op := range p.Operations {
						e := false
						for _, o := range perm.Operations {
							if op == o {
								e = true
								break
							}
						}
						if !e {
							perm.Operations = append(perm.Operations, op)
						}
					}
					break
				}
			}
			if !has {
				perms = append(perms, p)
			}
		}
	}
	return perms, nil
}
