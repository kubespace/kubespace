package manager

import (
	"errors"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/utils"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
	"regexp"
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
	user, err := u.Get(name)
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

type CreateUserRole struct {
	*UserRoleScope `json:",inline"`
	UserIds        []uint `json:"user_ids" form:"user_ids"`
}

type UserRoleScope struct {
	Scope      string           `json:"scope" form:"scope"`
	ScopeId    uint             `json:"scope_id" form:"scope_id"`
	ScopeRegex string           `json:"scope_regex" form:"scope_regex"`
	Role       string           `json:"role" form:"role"`
	SubScopes  []*UserRoleScope `json:"sub_scopes" form:"sub_scopes"`
}

func (r *UserRoleManager) CreateOrUpdate(createObj *CreateUserRole) error {
	for _, userId := range createObj.UserIds {
		if err := r.CreateOrUpdateUserRole(userId, types.RoleScopeRoot, types.RoleScopeRootId, createObj.UserRoleScope); err != nil {
			return err
		}
	}
	return nil
}

func (r *UserRoleManager) CreateOrUpdateUserRole(userId uint, parentScope string, parentScopeId uint, roleScope *UserRoleScope) error {
	var userRole types.UserRole
	if err := r.DB.First(&userRole, "user_id=? and scope=? and scope_id=? and parent_scope=? and parent_scope_id=?",
		userId, roleScope.Scope, roleScope.ScopeId, parentScope, parentScopeId).Error; err != nil {
		userRole = types.UserRole{
			UserId:        userId,
			Scope:         roleScope.Scope,
			ScopeId:       roleScope.ScopeId,
			ScopeRegex:    roleScope.ScopeRegex,
			ParentScope:   parentScope,
			ParentScopeId: parentScopeId,
			Role:          roleScope.Role,
			CreateTime:    time.Now(),
			UpdateTime:    time.Now(),
		}
		if err = r.DB.Create(&userRole).Error; err != nil {
			return err
		}
	} else {
		userRole.Role = roleScope.Role
		userRole.UpdateTime = time.Now()
		if err = r.DB.Save(&userRole).Error; err != nil {
			return err
		}
	}
	if err := r.DB.Delete(types.UserRole{}, "user_id = ? and parent_scope=? and parent_scope_id=?",
		userId, parentScope, parentScopeId).Error; err != nil {
		return err
	}
	for _, subScope := range roleScope.SubScopes {
		if err := r.CreateOrUpdateUserRole(userId, roleScope.Scope, roleScope.ScopeId, subScope); err != nil {
			return err
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

func (r *UserRoleManager) HasScopeRootRole(user *types.User, scope string, scopeId uint, scopeName string, role string) bool {
	return r.HasScopeRole(user, types.RoleScopeRoot, types.RoleScopeRootId, scope, scopeId, scopeName, role)
}

func (r *UserRoleManager) HasScopeRole(user *types.User, parentScope string, parentScopeId uint,
	scope string, scopeId uint, scopeName string, role string) bool {
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
	if scope == types.RoleScopePlatform {
		scopeId = 0
	}
	for _, scopeRole := range *user.Roles {
		if scopeRole.Scope == types.RoleScopePlatform && scopeRole.ScopeId == 0 && utils.Contains(roleSet, scopeRole.Role) {
			return true
		}
		if scopeRole.ParentScope != parentScope {
			continue
		}
		if scopeRole.ParentScopeId != parentScopeId {
			continue
		}
		if scopeRole.Scope != scope {
			continue
		}
		if scopeRole.ScopeRegex != "" {
			matched, err := regexp.MatchString(scopeRole.ScopeRegex, scopeName)
			if err != nil {
				klog.Errorf("regex %s match scope=%s scopeName=%s error: %s", scopeRole.ScopeRegex, scope, scopeName, err.Error())
				continue
			}
			if matched && utils.Contains(roleSet, scopeRole.Role) {
				return true
			}
			continue
		}
		if scopeRole.ScopeId == scopeId && utils.Contains(roleSet, scopeRole.Role) {
			return true
		}
	}
	return false
}
