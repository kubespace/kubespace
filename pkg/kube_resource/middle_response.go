package kube_resource

import "encoding/json"

const (
	RequestType = "request"
	WatchType   = "watch"
	TermType    = "exec"
	LogType     = "log"
)

type MiddleResponse struct {
	RequestId    string      `json:"request_id"`
	Data         interface{} `json:"data"`
	ResponseType string      `json:"res_type"`
}

func NewMiddleResponse(reqId, resType string, data interface{}) *MiddleResponse {
	return &MiddleResponse{
		RequestId:    reqId,
		Data:         data,
		ResponseType: resType,
	}
}

func UnserialzerMiddleResponse(d string) (*MiddleResponse, error) {
	var mr MiddleResponse
	err := json.Unmarshal([]byte(d), &mr)
	if err != nil {
		return nil, err
	}
	return &mr, nil
}

func (m *MiddleResponse) Serializer() ([]byte, error) {
	return json.Marshal(m.Data)
}

func (m *MiddleResponse) IsRequest() bool {
	return m.ResponseType == RequestType
}

func (m *MiddleResponse) IsWatch() bool {
	return m.ResponseType == WatchType
}

func (m *MiddleResponse) IsTerm() bool {
	return m.ResponseType == TermType
}

func (m *MiddleResponse) IsLog() bool {
	return m.ResponseType == LogType
}
