package project

import "gorm.io/gorm"

type ManagerProject struct {
	*gorm.DB
}

func NewManagerProject(db *gorm.DB) *ManagerProject {
	return &ManagerProject{DB: db}
}


