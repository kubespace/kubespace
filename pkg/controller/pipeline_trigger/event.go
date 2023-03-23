package pipeline_trigger

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/model/types"
)

func (p *PipelineTriggerController) eventLockKey(id uint) string {
	return fmt.Sprintf("pipeline_trigger_controller:event:%d", id)
}

// 流水线触发配置检查
func (p *PipelineTriggerController) eventCheck(obj interface{}) bool {
	event, ok := obj.(types.PipelineTriggerEvent)
	if !ok {
		return false
	}
	if locked, _ := p.lock.Locked(p.eventLockKey(event.ID)); locked {
		// 正在处理该流水线触发事件
		return false
	}
	if event.Status != types.PipelineTriggerEventStatusNew {
		return false
	}
	return true
}

// 流水线触发事件处理
func (p *PipelineTriggerController) eventHandle(obj interface{}) error {
	event := obj.(types.PipelineTriggerEvent)
	// 对流水线配置处理加锁，保证只有一个goroutinue执行
	if ok, _ := p.lock.Acquire(p.eventLockKey(event.ID)); !ok {
		return nil
	}
	// 执行完成释放锁
	defer p.lock.Release(p.eventLockKey(event.ID))

	pipeline, err := p.models.PipelineManager.Get(event.PipelineId)
	if err != nil {
		return err
	}
	pipelineWorkspace, err := p.models.PipelineWorkspaceManager.Get(pipeline.WorkspaceId)
	if err != nil {
		return err
	}
	if pipelineWorkspace.Type == types.WorkspaceTypeCode {

	}
	return nil
}
