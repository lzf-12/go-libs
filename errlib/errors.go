package errlib

import (
	"fmt"
)

type CustomError struct {
	Code    ErrorCode
	Message string
	Trace   string
	Err     error
}

func New(code ErrorCode, message string, err error) *CustomError {
	return &CustomError{
		Code:    code,
		Message: message,
		Trace:   GetStackTrace(),
		Err:     err,
	}
}

func (e *CustomError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Trace)
}

func (e *CustomError) Unwrap() error {
	return e.Err
}

func (e *CustomError) WithMessage(message string) *CustomError {
	e.Message = message
	return e
}

func Wrap(err error, code ErrorCode, message string) *CustomError {
	return New(code, message, err)
}
