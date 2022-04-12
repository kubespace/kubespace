package manager

import (
	"errors"
	"github.com/kubespace/kubespace/pkg/model/types"
	"gorm.io/gorm"
	"time"
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

func (u *UserManager) GetById(id uint) (*types.User, error) {
	var user types.User
	if err := u.DB.First(&user, "id=?", id).Error; err != nil {
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
	//	roleObj, err := u.role.GetByName(role)
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

type UserRoleManager struct {
	DB          *gorm.DB
	UserManager *UserManager
}

func NewUserRoleManager(db *gorm.DB, user *UserManager) *UserRoleManager {
	return &UserRoleManager{DB: db, UserManager: user}
}

func (r *UserRoleManager) List(scope string, scopeId uint) ([]*types.UserRole, error) {
	var roles []types.UserRole
	if err := r.DB.Find(&roles, "scope=? and scope_id=?", scope, scopeId).Error; err != nil {
		return nil, err
	}
	var resRoles []*types.UserRole
	for i, role := range roles {
		user, err := r.UserManager.GetById(role.UserId)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, err
			}
		} else {
			roles[i].UserName = user.Name
		}
		resRoles = append(resRoles, &roles[i])
	}
	return resRoles, nil
}

func (r *UserRoleManager) CreateOrUpdate(scope string, scopeId uint, userIds []uint, role string) error {
	for _, userId := range userIds {
		var userRole types.UserRole
		if err := r.DB.First(&userRole, "user_id=? and scope=? and scope_id=?", userId, scope, scopeId).Error; err != nil {
			userRole = types.UserRole{
				UserId:     userId,
				Scope:      scope,
				ScopeId:    scopeId,
				Role:       role,
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
			}
			if err = r.DB.Create(&userRole).Error; err != nil {
				return err
			}
		} else {
			userRole.Role = role
			userRole.UpdateTime = time.Now()
			if err = r.DB.Save(&userRole).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *UserRoleManager) Delete(id uint) error {
	var userRole types.UserRole
	if err := r.DB.First(&userRole, "id=?", id).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	} else {
		if err = r.DB.Delete(&userRole).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *UserRoleManager) GetUserRoles(userId uint) ([]types.UserRole, error) {
	var userRoles []types.UserRole
	if err := r.DB.Find(&userRoles, "user_id=?", userId).Error; err != nil {
		return nil, err
	}
	return userRoles, nil
}
