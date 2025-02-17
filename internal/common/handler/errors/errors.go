package errors

import (
	"errors"
	"fmt"

	"github.com/PIGcanstudy/gorder/common/consts"
)

type Error struct {
	code int
	msg  string
	err  error
}

// 返回错误码对应的信息
func (e *Error) Error() string {
	var msg string
	if e.msg != "" {
		msg = e.msg
	}
	msg = consts.ErrMsg[e.code]
	return msg + " -> " + e.err.Error()
}

func New(code int) error {
	return &Error{
		code: code,
	}
}

func NewWithError(code int, err error) error {
	if err == nil {
		return New(code)
	}
	return &Error{
		code: code,
		err:  err,
	}
}

func NewWithMsgf(code int, format string, args ...any) error {
	return &Error{
		code: code,
		msg:  fmt.Sprintf(format, args...),
	}
}

// 返回错误对应的错误码
func Errno(err error) int {
	if err == nil {
		return consts.ErrnoSuccess
	}
	targetError := &Error{}
	// 判断err是否是我们定义的Error类型
	if errors.As(err, &targetError) {
		return targetError.code
	}
	return -1
}

func Output(err error) (int, string) {
	if err == nil {
		return consts.ErrnoSuccess, consts.ErrMsg[consts.ErrnoSuccess]
	}
	errno := Errno(err)
	if errno == -1 {
		return consts.ErrnoUnknownError, err.Error()
	}
	return errno, err.Error()
}
