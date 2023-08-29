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

func (e *Error) String() string {
	return fmt.Sprintf("%s:%v", e.code, e.err)
}

func (e *Error) Code() string {
	return e.code
}

type options struct {
	// 是否覆盖错误码
	overlap bool
}

type optionFunc func(o *options)

var Overlap = func(o *options) {
	o.overlap = true
}

func New(code string, e interface{}, opfs ...optionFunc) *Error {
	o := options{}
	for _, opf := range opfs {
		opf(&o)
	}
	var err error

	switch e := e.(type) {
	case *Error:
		if !o.overlap {
			return e
		}
		// 覆盖当前错误码
		return &Error{code: code, err: fmt.Errorf("%s", e)}
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
