package user

import (
	"errors"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/utils"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
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

func (u *UserManager) GetByName(name string) (*types.User, error) {
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

type UserListCondition struct {
	Ids []uint `json:"ids"`
}

func (u *UserManager) List(cond UserListCondition) ([]*types.User, error) {
	var users []*types.User
	tx := u.DB
	if len(cond.Ids) > 0 {
		tx = tx.Where("id in ?", cond.Ids)
	}
	if err := tx.Order("name").Find(&users).Error; err != nil {
		return nil, err
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
	user, err := u.GetByName(name)
	if err != nil {
		return err
	}
	if err = u.DB.Delete(types.UserRole{}, "user_id = ?", user.ID).Error; err != nil {
		return err
	}
	if err = u.DB.Delete(types.User{}, "name = ?", name).Error; err != nil {
		return err
	}
	return nil
}

func (u *UserManager) Permissions(user *types.User) ([]types.Permission, error) {
	var perms []types.Permission
	return perms, nil
}

type UserRoleManager struct {
	DB          *gorm.DB
	UserManager *UserManager
}

func NewUserRoleManager(db *gorm.DB, user *UserManager) *UserRoleManager {
	return &UserRoleManager{DB: db, UserManager: user}
}

func (r *UserRoleManager) GetById(id uint) (*types.UserRole, error) {
	var userRole types.UserRole
	if err := r.DB.First(&userRole, "id=?", id).Error; err != nil {
		return nil, err
	}
	return &userRole, nil
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

func (r *UserRoleManager) HasScopeRole(user *types.User, scope string, scopeId uint, role string) bool {
	if user.IsSuper {
		return true
	}
	if user.Roles == nil {
		roles, err := r.GetUserRoles(user.ID)
		if err != nil {
			klog.Errorf("get user id=%d error: %s", user.ID, err.Error())
			return false
		}
		user.Roles = &roles
	}
	roleSetsMap := map[string][]string{
		types.RoleTypeViewer: {types.RoleTypeViewer, types.RoleTypeEditor, types.RoleTypeAdmin},
		types.RoleTypeEditor: {types.RoleTypeEditor, types.RoleTypeAdmin},
		types.RoleTypeAdmin:  {types.RoleTypeAdmin},
	}
	roleSet, ok := roleSetsMap[role]
	if !ok {
		klog.Errorf("not found role %s", role)
		return false
	}
	for _, scopeRole := range *user.Roles {
		if scopeRole.Scope == scope && scopeRole.ScopeId == scopeId && utils.Contains(roleSet, scopeRole.Role) {
			return true
		}
		if scopeRole.Scope == types.ScopePlatform && scopeRole.ScopeId == 0 && utils.Contains(roleSet, scopeRole.Role) {
			return true
		}
	}
	return false
}
