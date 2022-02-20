package manager

import (
	"github.com/kubespace/kubespace/pkg/model/types"
	"gorm.io/gorm"
)

type UserManager struct {
	//CommonManager
	//role *RoleManager
	DB *gorm.DB
}

func NewUserManager(db *gorm.DB) *UserManager {
	return &UserManager{
		DB: db,
	}
}
//
//func (u *UserManager) parseToStore(user *types.User) (*types.UserStore, error) {
//	roleBytes, err := json.Marshal(user.Roles)
//	if err != nil {
//		return nil, err
//	}
//	userStore := &types.UserStore{
//		Name:      user.Name,
//		Email:     user.Email,
//		Password:  user.Password,
//		Status:    user.Status,
//		IsSuper:   strconv.FormatBool(user.IsSuper),
//		Roles:     string(roleBytes),
//		LastLogin: user.LastLogin,
//	}
//	return userStore, nil
//}
//
//func (u *UserManager) parseToUser(userStore *types.UserStore) (*types.User, error) {
//	var roles []string
//	json.Unmarshal([]byte(userStore.Roles), &roles)
//
//	user := &types.User{
//		Name:      userStore.Name,
//		Email:     userStore.Email,
//		Common:    userStore.Common,
//		Password:  userStore.Password,
//		Status:    userStore.Status,
//		IsSuper:   utils.ParseBool(userStore.IsSuper),
//		LastLogin: userStore.LastLogin,
//		Roles:     roles,
//	}
//	return user, nil
//}

func (u *UserManager) Get(name string) (*types.User, error) {
	var user types.User
	if err := u.DB.First(&user, "name=?", name).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *UserManager) List(filters map[string]interface{}) ([]types.User, error) {
	//dList, err := u.CommonManager.List(filters)
	//if err != nil {
	//	return nil, err
	//}
	//jsonBody, err := json.Marshal(dList)
	//if err != nil {
	//	return nil, err
	//}
	//var userStores []*types.UserStore
	//
	//if err := json.Unmarshal(jsonBody, &userStores); err != nil {
	//	return nil, err
	//}
	//
	//var users []*types.User
	//for _, rs := range userStores {
	//	user, err := u.parseToUser(rs)
	//	if err != nil {
	//		return nil, err
	//	}
	//	users = append(users, user)
	//}
	//return users, nil

	var users []types.User
	result := u.DB.Where(filters).Order("name").Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

func (u *UserManager) Update(user *types.User) error {
	if err := u.DB.Save(user).Error; err != nil {
		return err
	}
	return nil
}

func (u *UserManager) Create(user *types.User) error {
	if err := u.DB.Create(user).Error; err != nil {
		return err
	}
	return nil
}

func (u *UserManager) Delete(name string) error {
	if err := u.DB.Delete(types.User{}, "name = ?", name).Error; err != nil {
		return err
	}
	return nil
}

func (u *UserManager) Permissions(user *types.User) ([]types.Permission, error) {
	var perms []types.Permission

	//for _, role := range user.Roles {
	//	roleObj, err := u.role.Get(role)
	//	if err != nil {
	//		return nil, err
	//	}
	//	for _, p := range roleObj.Permissions {
	//		has := false
	//		for _, perm := range perms {
	//			if p.Scope == perm.Scope && p.Object == perm.Object {
	//				has = true
	//				for _, op := range p.Operations {
	//					e := false
	//					for _, o := range perm.Operations {
	//						if op == o {
	//							e = true
	//							break
	//						}
	//					}
	//					if !e {
	//						perm.Operations = append(perm.Operations, op)
	//					}
	//				}
	//				break
	//			}
	//		}
	//		if !has {
	//			perms = append(perms, p)
	//		}
	//	}
	//}
	return perms, nil
}
