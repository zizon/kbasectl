package panichain

import (
	"fmt"
	"runtime"
)

func Propogate(err error) {
	if err == nil {
		return
	}
	buf := make([]byte, 1024)
	runtime.Stack(buf, false)

	panic(fmt.Errorf(
		"reason: %w stacktrace:\n%s",
		err,
		buf,
	))
}

func Catch(handler func(error) error) {
	if any := recover(); any != nil {
		if err, ok := any.(error); ok {
			Propogate(handler(err))
			return
		}
	}
}
