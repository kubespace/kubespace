package pipeline

import (
	"errors"
	"github.com/kubespace/kubespace/pkg/model/types"
	"gorm.io/gorm"
)

type Release struct {
	DB *gorm.DB
}

func NewReleaseManager(db *gorm.DB) *Release {
	return &Release{DB: db}
}

func (l *Release) GetLatestRelease(workspaceId uint) (*types.PipelineWorkspaceRelease, error) {
	var release types.PipelineWorkspaceRelease
	if err := l.DB.Last(&release, "workspace_id = ?", workspaceId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &release, nil
}

func (l *Release) ExistsRelease(workspaceId uint, version string) (bool, error) {
	var cnt int64
	if err := l.DB.Model(&types.PipelineWorkspaceRelease{}).Where("workspace_id = ? and release_version = ?", workspaceId, version).Count(&cnt).Error; err != nil {
		return false, err
	}
	if cnt > 0 {
		return true, nil
	}
	return false, nil
}
