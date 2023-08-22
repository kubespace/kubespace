package errors

import (
	"fmt"
)

// Error 自定义错误码数据结构
type Error struct {
	code string
	err  error
}

func (e *Error) Error() string {
	return e.err.Error()
}

func (e *Error) Code() string {
	return e.code
}

func New(code string, e interface{}) *Error {
	var err error

	switch e := e.(type) {
	case *Error:
		return e
	case error:
		err = e
	default:
		err = fmt.Errorf("%v", e)
	}
	return &Error{code: code, err: err}
}

// IsCode 判断错误是否属于该错误码
func IsCode(err error, code string) bool {
	if e, ok := err.(*Error); ok {
		if e.code == code {
			return true
		}
	}
	return false
}
