package project

import (
	"gorm.io/gorm"
)

type AppChartManager struct {
	*gorm.DB
}

func NewAppChartManager(db *gorm.DB) *AppChartManager {
	return &AppChartManager{DB: db}
}

//func (c *AppChartManager) Create(path, chartFilePath string) (*types.AppVersionChart, error) {
//
//}
