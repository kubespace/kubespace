package kube_resource

import (
	"encoding/json"
	"github.com/kubespace/kubespace/pkg/utils"
)

type MiddleRequest struct {
	Cluster   string      `json:"cluster"`
	RequestId string      `json:"request_id"`
	Resource  string      `json:"resource"`
	Action    string      `json:"action"`
	Params    interface{} `json:"params"`
	Timeout   int64
}

func NewMiddleRequest(cluster, resType, action string, params interface{}, timeout int64) *MiddleRequest {
	requestId := utils.CreateUUID()
	if timeout <= 0 {
		timeout = 120
	}
	return &MiddleRequest{
		Cluster:   cluster,
		RequestId: requestId,
		Resource:  resType,
		Action:    action,
		Params:    params,
		Timeout:   timeout,
	}
}

func (r *MiddleRequest) Serializer() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"cluster":    r.Cluster,
		"request_id": r.RequestId,
		"resource":   r.Resource,
		"action":     r.Action,
		"params":     r.Params,
	})
}

func UnserializerMiddleRequest(data string) (*MiddleRequest, error) {
	var mr MiddleRequest
	err := json.Unmarshal([]byte(data), &mr)
	if err != nil {
		return nil, err
	}
	return &mr, nil
}
