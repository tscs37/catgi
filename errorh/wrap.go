package errorh

import "fmt"

func CatchPanic(f func() error) (err error) {
	defer func() {
		if pErr := recover(); pErr != nil {
			var ok bool
			err, ok = pErr.(error)
			if !ok {
				Panicf("Panic on non-error: %#v", err)
			}
		}
	}()
	return f()
}

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func Panicf(message string, variables ...interface{}) {
	panic(fmt.Sprintf(message, variables...))
}
