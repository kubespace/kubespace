package manager

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/utils"
	"math"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type Pagination struct {
	PageSize int64 `json:"page_size"`
	PageNo   int64 `json:"page_no"`
	Pages    int64 `json:"pages"`
	Records  int64 `json:"records"`
	Offset   int64 `json:"-"`
	Limit    int64 `json:"-"`
}

type PaginationCondition struct {
	PageSize int64  `form:"page_size,default=20"` // page_size=0表示不分页
	PageNo   int64  `form:"page_no,default=1"`
	OrderBy  string `form:"order_by"` // order_by=field表示升序，order_by=-field表示降序
}

// NewPagination 创建分页对象
func NewPagination(pageSize int64, pageNo int64, records int64) *Pagination {
	pagination := &Pagination{
		PageSize: utils.MaxInt64(pageSize, 0),
		PageNo:   utils.MaxInt64(pageNo, 1),
		Records:  records,
	}
	if pagination.PageSize == 0 {
		pagination.Limit = pagination.Records
	} else {
		pagination.Pages = int64(math.Ceil(float64(pagination.Records) / float64(pagination.PageSize)))
		pagination.Offset = utils.MinInt64(pagination.Records, (pagination.PageNo-1)*pagination.PageSize)
		pagination.Limit = utils.MinInt64(pagination.PageSize, utils.MaxInt64(pagination.Records-pagination.Offset, 0))
	}
	return pagination
}

// NewPaginationFromDb 通过查询数据库创建分页对象
func NewPaginationFromDb(tx *gorm.DB, modelType schema.Tabler, modelList interface{}, cond PaginationCondition) (*Pagination, error) {
	tx = tx.Model(modelType)
	var totalCount int64
	tx.Count(&totalCount).Debug()
	pagination := NewPagination(cond.PageSize, cond.PageNo, totalCount)
	if pagination.Limit == 0 {
		return pagination, nil
	}
	if cond.OrderBy != "" {
		if strings.HasPrefix(cond.OrderBy, "-") {
			cond.OrderBy = fmt.Sprintf("`%s` DESC", cond.OrderBy[1:])
		} else {
			cond.OrderBy = fmt.Sprintf("`%s` ASC", cond.OrderBy)
		}
		tx.Order(cond.OrderBy)
	}
	if pagination.PageSize > 0 {
		tx.Offset(int(pagination.Offset)).Limit(int(pagination.Limit))
	}
	if err := tx.Debug().Find(modelList).Error; err != nil {
		return nil, err
	}
	return pagination, nil
}
