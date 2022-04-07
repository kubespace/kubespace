package pipeline

import (
	"github.com/kubespace/kubespace/pkg/model/types"
	"gorm.io/gorm"
	"time"
)

type JobLog struct {
	DB *gorm.DB
}

func NewJobLogManager(db *gorm.DB) *JobLog {
	return &JobLog{DB: db}
}

func (l *JobLog) UpdateLog(jobId uint, log string) error {
	var cnt int64
	if err := l.DB.Model(&types.PipelineRunJobLog{}).Where("job_run_id", jobId).Count(&cnt).Error; err != nil {
		return err
	}
	if cnt == 0 {
		log := types.PipelineRunJobLog{
			JobRunId:   jobId,
			Logs:       log,
			CreateTime: time.Now(),
			UpdateTime: time.Now(),
		}
		if err := l.DB.Create(&log).Error; err != nil {
			return err
		}
	} else {
		if err := l.DB.Model(types.PipelineRunJobLog{}).Where("job_run_id", jobId).Updates(types.PipelineRunJobLog{Logs: log}).Error; err != nil {
			return err
		}
	}
	return nil
}
