package utils

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kubespace/kubespace/pkg/core/code"
)

type Response struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

func (r *Response) IsSuccess() bool {
	return r.Code == code.Success
}

func (r *Response) Scan(value any) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to convert to bytes:", value))
	}
	err := json.Unmarshal(bytes, r)
	if err != nil {
		return fmt.Errorf("failed to unmarshal bytes: %s", string(bytes))
	}
	return nil
}

// Value return json value, implement driver.Valuer interface
func (r Response) Value() (driver.Value, error) {
	bytes, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	return string(bytes), nil
}

type WatchResponse struct {
	Event    string `json:"event"`
	Obj      string `json:"obj"`
	Resource any    `json:"resource"`
}
