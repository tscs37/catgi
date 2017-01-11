package utils

import (
	"time"
)

// WaitFor checks the predicate every 10 milliseconds and if true returns
func WaitFor(pred func() bool) {
	for !pred() {
		time.Sleep(10 * time.Millisecond)
	}
}
