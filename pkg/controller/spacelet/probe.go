package spacelet

import (
	"fmt"
	"github.com/kubespace/kubespace/pkg/model/types"
	spaceletservice "github.com/kubespace/kubespace/pkg/service/spacelet"
	spacelet "github.com/kubespace/kubespace/pkg/spacelet"
	"k8s.io/klog/v2"
	"time"
)

func (s *SpaceletController) probeLockKey(id uint) string {
	return fmt.Sprintf("spacelet:probe:%d", id)
}

func (s *SpaceletController) probeCheck(obj interface{}) bool {
	spaceletObj, ok := obj.(types.Spacelet)
	if !ok {
		return false
	}
	if locked, _ := s.lock.Locked(s.probeLockKey(spaceletObj.ID)); locked {
		return false
	}
	return true
}

// 定时探测spacelet节点是否存活
func (s *SpaceletController) probe(obj interface{}) error {
	spaceletObj := obj.(types.Spacelet)
	status := s.status(&spaceletObj)
	if spaceletObj.Status != status {
		klog.Infof("spacelet host=%s ip=%s stauts=%s", spaceletObj.Hostname, spaceletObj.HostIp, status)
		return s.models.SpaceletManager.Update(spaceletObj.ID, &types.Spacelet{
			Status:     status,
			UpdateTime: time.Now(),
		})
	}
	return nil
}

func (s *SpaceletController) status(spaceletObj *types.Spacelet) string {
	spaceletClient, err := spaceletservice.NewClient(spaceletObj)
	if err != nil {
		return types.SpaceletStatusOffline
	}
	execResp, err := spaceletClient.Exec(&spacelet.ExecRequest{
		Command: "echo",
	})
	if err != nil {
		klog.Errorf("spacelet host=%s ip=%s exec error: %s", spaceletObj.Hostname, spaceletObj.HostIp, err.Error())
		return types.SpaceletStatusOffline
	}
	if execResp.Status == 1 {
		klog.Errorf("spacelet host=%s ip=%s exec status error: %s", spaceletObj.Hostname, spaceletObj.HostIp, execResp.Stderr)
		return types.SpaceletStatusOffline
	}
	return types.SpaceletStatusOnline
}
