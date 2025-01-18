package errlib

import (
	"fmt"
	"runtime"
)

func GetStackTrace() string {
	var stackTrace string
	for i := 2; ; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		stackTrace += fmt.Sprintf("at %s:%d\n", file, line)
	}
	return stackTrace
}
