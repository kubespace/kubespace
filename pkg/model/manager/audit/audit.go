package audit

import (
	"github.com/kubespace/kubespace/pkg/model/manager"
	"github.com/kubespace/kubespace/pkg/model/types"
	"gorm.io/gorm"
)

type AuditOperateManager struct {
	*gorm.DB
}

func NewAuditOperateManager(db *gorm.DB) *AuditOperateManager {
	return &AuditOperateManager{DB: db}
}

func (a *AuditOperateManager) Create(ao *types.AuditOperate) error {
	return a.DB.Create(ao).Error
}

type AuditOperateListCondition struct {
	manager.PaginationCondition
	Scope     string `json:"scope" form:"scope"`
	ScopeId   uint   `json:"scope_id" form:"scope_id"`
	Namespace string `json:"namespace" form:"namespace"`
	FuzzyName string `json:"fuzzy_name" form:"fuzzy_name"`
}

func (a *AuditOperateManager) List(cond *AuditOperateListCondition) ([]*types.AuditOperate, *manager.Pagination, error) {
	var aos []*types.AuditOperate
	tx := a.DB
	if cond.Scope != "" {
		tx = tx.Where("scope = ? and scope_id = ?", cond.Scope, cond.ScopeId)
	}
	if cond.Namespace != "" {
		tx = tx.Where("namespace = ?", cond.Namespace)
	}
	if cond.FuzzyName != "" {
		tx = tx.Where("resource_name like ?", "%"+cond.FuzzyName+"%")
	}
	if cond.OrderBy == "" {
		cond.PaginationCondition.OrderBy = "-create_time"
	}

	page, err := manager.NewPaginationFromDb(tx, &types.AuditOperate{}, &aos, cond.PaginationCondition)
	if err != nil {
		return nil, nil, err
	}
	return aos, page, err
}
